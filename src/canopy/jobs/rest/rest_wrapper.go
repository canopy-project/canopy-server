// Copyright 2015 Canopy Services, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
    "canopy/canolog"
    "canopy/config"
    "canopy/datalayer"
    "canopy/jobqueue"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
)

// CanopyRestAuthTypeEnum is the type of authentication used in a request
type CanopyRestAuthTypeEnum int
const (
    // Request did not include any authentication
    CANOPY_REST_AUTH_NONE = iota

    // Request included HTTP BASIC authentication for a user account
    CANOPY_REST_AUTH_BASIC

    // Request included HTTP BASIC authentication for a device
    CANOPY_REST_AUTH_DEVICE_BASIC

    // Request included a session cookie
    CANOPY_REST_AUTH_SESSION
)

type CanopyRestInfo struct {
    AuthType CanopyRestAuthTypeEnum
    Account datalayer.Account
    BodyObj map[string]interface{}
    Conn datalayer.Connection
    Config config.Config
    Device datalayer.Device
    Cookies map[string]string
    PigeonClient jobqueue.Client
    URLVars map[string]string
    UserCtx map[string]interface{}
}

func parseBasicAuth(authHeader string) (username string, password string, err error) {
    if len(authHeader) == 0 {
        return "", "", errors.New("Authorization header not set")
    }
    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 {
        return "", "", errors.New("Authentication header malformed")
    }
    if parts[0] != "Basic" {
        return "", "", errors.New("Expected basic authentication")
    }
    encodedVal := parts[1]
    decodedVal, err := base64.StdEncoding.DecodeString(encodedVal)
    if err != nil {
        return "", "", errors.New("Authentication header malformed")
    }
    parts = strings.Split(string(decodedVal), ":")
    if len(parts) != 2 {
        return "", "", errors.New("Authentication header malformed")
    }
    return parts[0], parts[1], nil
}

type RestJobHandler func(
        info *CanopyRestInfo,
        req jobqueue.Request,
        resp jobqueue.Response)


// Wrapper for handling pigeon requests that originated from
// CanopyRestJobForwarder
func RestJobWrapper(handler RestJobHandler) jobqueue.HandlerFunc {
    return func(jobKey string, userCtx map[string]interface{}, req jobqueue.Request, resp jobqueue.Response) {
        // This expects to recieve the following over the wire from the Pigeon
        // client:
        //  {
        //      "url-vars" : map[string]string,
        //      "auth-header" : string,
        //      "cookie-username" : string,
        //      "http-body" : string,
        //  }
        var ok bool

        canolog.Info("Handling job", jobKey)
        info := &CanopyRestInfo{ }
        body := req.Body()
        canolog.Info("Request:", body)

        // Get URL vars from job request
        info.URLVars, ok = body["url-vars"].(map[string]string)
        if !ok {
            resp.SetError(fmt.Errorf("Expected map[string]string for 'url-vars'"))
            return
        }

        // Get DB Connection from userCtx
        info.Conn, ok = userCtx["db-conn"].(datalayer.Connection)
        conn := info.Conn
        if !ok {
            resp.SetError(fmt.Errorf("Expected datalayer.Connection for 'db-conn'"))
            return
        }

        // Check for BASIC AUTH
        authHeader, ok := body["auth-header"].(string)
        if !ok {
            resp.SetError(fmt.Errorf("Expected string for 'auth-header'"))
            return
        }
        username_string, password, err := parseBasicAuth(authHeader)
        if err == nil {
            // was a UUID provided?
            if len(username_string) == 36 {
                device, err := info.Conn.LookupDeviceByStringID(username_string)
                if err != nil {
                    resp.SetError(fmt.Errorf("Incorrect username or password"))
                    //w.WriteHeader(http.StatusUnauthorized)
                    //fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
                    return
                }
                
                if device.SecretKey() != password {
                    resp.SetError(fmt.Errorf("Incorrect username or password"))
                    //w.WriteHeader(http.StatusUnauthorized)
                    //fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
                    return
                }

                info.AuthType = CANOPY_REST_AUTH_DEVICE_BASIC
                info.Device = device

                // update last_seen for this device
                err = device.UpdateLastActivityTime(nil)
                if err != nil {
                    resp.SetError(fmt.Errorf("Updating last seen time: %s", err.Error()))
                    //rest_errors.NewInternalServerError("Updating last seen time").WriteTo(w)
                    return
                }
                canolog.Info("Device BASIC auth provided")
            } else {
                // otherwise, assume user account username/password provided
                acct, err := conn.LookupAccountVerifyPassword(username_string, password)
                if err != nil {
                    if err == datalayer.InvalidPasswordError {
                        resp.SetError(fmt.Errorf("Incorrect username or password"))
                        //w.WriteHeader(http.StatusUnauthorized)
                        //fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
                        return
                    } else {
                        resp.SetError(fmt.Errorf("Account lookup failed"))
                        //w.WriteHeader(http.StatusInternalServerError);
                        //fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
                        return
                    }
                }
                
                canolog.Info("Basic auth provided")
                info.AuthType = CANOPY_REST_AUTH_BASIC
                info.Account = acct
            }
        }

        // Check for session-based AUTH
        info.Cookies = make(map[string]string)
        info.Cookies["username"], ok = body["cookie-username"].(string)
        if !ok {
            resp.SetError(fmt.Errorf("Expected string for 'cookie-username'"))
            return
        }

        username, ok := info.Cookies["username"]
        if ok && username != "" {
            canolog.Info("Looking up account: ", username)
            acct, err := conn.LookupAccount(username)
            if err != nil {
                // TODO: Do this Logout logic
                //info.Session.Values["logged_in_username"] = ""
                //info.Session.Save(r, w)
                //w.WriteHeader(http.StatusInternalServerError);
                //fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
                resp.SetError(fmt.Errorf("Account lookup failed"))
                return
            }

            canolog.Info("Session auth provided")
            info.AuthType = CANOPY_REST_AUTH_SESSION
            info.Account = acct
        }

        httpBody, ok := body["http-body"].(string)
        if !ok {
            resp.SetError(fmt.Errorf("Expected string for 'http-body'"))
            return
        }

        // Decode httpBody JSON
        var bodyObj map[string]interface{}
        if httpBody != "" {
            decoder := json.NewDecoder(strings.NewReader(httpBody))
            err := decoder.Decode(&bodyObj)
            if err != nil {
                resp.SetError(fmt.Errorf("JSON decode failed: %s ", err.Error()))
                return
            }
        }
        info.BodyObj = bodyObj

        // Call the wrapped handler.
        // The wrapped handler should set response in <resp> using:
        //      resp.SetBody()
        //      resp.SetError()
        handler(info, req, resp)

        canolog.Info("Response: ", resp.Body())
        canolog.Info("Response (Error): ", resp.Error())
    }
}
