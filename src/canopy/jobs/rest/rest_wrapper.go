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
    "bytes"
    "canopy/canolog"
    "canopy/config"
    "canopy/datalayer"
    "canopy/jobqueue"
    "canopy/mail"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "runtime"
    "strings"
)

type RestJobHandler func(reqInfo *RestRequestInfo, sideeffect *RestSideEffects) (respJson map[string]interface{}, err RestError) 

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

type RestRequestInfo struct {
    AuthType CanopyRestAuthTypeEnum
    Account datalayer.Account
    BodyObj map[string]interface{}
    Conn datalayer.Connection
    Config config.Config
    Device datalayer.Device
    Cookies map[string]string
    PigeonOutbox jobqueue.Outbox
    URLVars map[string]string
    UserCtx map[string]interface{}
}

func parseBasicAuth(authHeader []string) (username string, password string, err error) {
    if len(authHeader) == 0 {
        return "", "", errors.New("Authorization header not set")
    }
    parts := strings.SplitN(authHeader[0], " ", 2)
    if len(parts) != 2 {
        return "", "", errors.New("Authentication header malformed")
    }
    if !strings.EqualFold(parts[0], "Basic") {
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

func RestSetError(resp jobqueue.Response, err RestError) {
    resp.SetBody(map[string]interface{}{
        "http-status" : err.StatusCode(),
        "http-body" : err.ResponseBody(),
    })
}

func RestSetErrorClearCookies(resp jobqueue.Response, err RestError) {
    resp.SetBody(map[string]interface{}{
        "http-status" : err.StatusCode(),
        "http-body" : err.ResponseBody(),
        "clear-cookies" : []string{"logged_in_username"},
    })
}

// Wrapper for handling pigeon requests that originated from
// CanopyRestJobForwarder
func RestJobWrapper(handler RestJobHandler) jobqueue.HandlerFunc {
    return func(jobKey string, userCtxItf interface{}, req jobqueue.Request, resp jobqueue.Response) {
        // This expects to recieve the following over the wire from the Pigeon
        // client:
        //  {
        //      "url-vars" : map[string]string,
        //      "auth-header" : string,
        //      "cookie-username" : string,
        //      "http-body" : string,
        //  }
        //
        // This sends the following response to the Pigeon client:
        //  {
        //      "http-status" : int,
        //      "http-body" : string,
        //      "clear-cookies" : []string,
        //      "set-cookies" : map[string]string,
        //  }
        var ok bool

        defer func() {
            // Catch exceptions and return callstack
            r := recover()
            if r != nil {
                var buf [4096]byte
                runtime.Stack(buf[:], false)
                n := bytes.Index(buf[:], []byte{0})
                canolog.Error(string(buf[:n]))
                RestSetError(resp, InternalServerError(fmt.Sprint("Crash: ", string(buf[:n]))))
            }
        }()

        canolog.Info("Handling job", jobKey)
        info := &RestRequestInfo{ }
        body := req.Body()
        canolog.Info("Request:", body)

        // Get URL vars from job request
        info.URLVars, ok = body["url-vars"].(map[string]string)
        if !ok {
            RestSetError(resp, InternalServerError("Expected map[string]string for 'url-vars'").Log())
            return
        }

        userCtx, ok := userCtxItf.(map[string]interface{})
        if !ok {
            RestSetError(resp, InternalServerError("Expected map[string]interface{} for userCtx").Log())
            return
        }

        // Get DB Connection from userCtx
        info.Conn, ok = userCtx["db-conn"].(datalayer.Connection)
        conn := info.Conn
        if !ok {
            RestSetError(resp, InternalServerError("Expected datalayer.Connection for 'db-conn'").Log())
            return
        }

        // Get Config from userCtx
        info.Config, ok = userCtx["cfg"].(config.Config)
        if !ok {
            RestSetError(resp, InternalServerError("Expected config.Config for 'cfg'").Log())
            return
        }

        // Get MailClient from userCtx
        mailer, ok := userCtx["mailer"].(mail.MailClient)
        if !ok {
            RestSetError(resp, InternalServerError("Expected MailClient for 'mailer'").Log())
            return
        }

        // Check for BASIC AUTH
        authHeader, ok := body["auth-header"].([]string)
        if !ok {
            RestSetError(resp, InternalServerError("Expected []string for 'auth-header'").Log())
            return
        }
        username_string, password, err := parseBasicAuth(authHeader)
        if err == nil {
            // was a UUID provided?
            if len(username_string) == 36 {
                device, err := info.Conn.LookupDeviceByStringID(username_string)
                if err != nil {
                    RestSetError(resp, IncorrectUsernameOrPasswordError().Log())
                    return
                }
                
                if device.SecretKey() != password {
                    RestSetError(resp, IncorrectUsernameOrPasswordError().Log())
                    return
                }

                info.AuthType = CANOPY_REST_AUTH_DEVICE_BASIC
                info.Device = device

                // update last_seen for this device
                err = device.UpdateLastActivityTime(nil)
                if err != nil {
                    RestSetError(resp, InternalServerError("Updating last seen time " + err.Error()).Log())
                    return
                }
                canolog.Info("Device BASIC auth provided")
            } else {
                // otherwise, assume user account username/password provided
                acct, err := conn.LookupAccountVerifyPassword(username_string, password)
                if err != nil {
                    if err == datalayer.InvalidPasswordError {
                        RestSetError(resp, IncorrectUsernameOrPasswordError().Log())
                        return
                    } else {
                        RestSetError(resp, InternalServerError("Account lookup failed").Log())
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
            RestSetError(resp, InternalServerError("Expected string for 'cookie-username'").Log())
            return
        }

        username, ok := info.Cookies["username"]
        if ok && username != "" {
            canolog.Info("Looking up account: ", username)
            acct, err := conn.LookupAccount(username)
            if err != nil {
                // TODO: Handle clear cookie logic on client side as well
                RestSetErrorClearCookies(resp, InternalServerError("Account lookup failed").Log())
                return
            }

            canolog.Info("Session auth provided")
            info.AuthType = CANOPY_REST_AUTH_SESSION
            info.Account = acct
        }

        httpBody, ok := body["http-body"].(string)
        if !ok {
            RestSetError(resp, InternalServerError("Expected string for 'http-body'").Log())
            return
        }

        // Decode httpBody JSON
        var bodyObj map[string]interface{}
        if httpBody != "" {
            decoder := json.NewDecoder(strings.NewReader(httpBody))
            err := decoder.Decode(&bodyObj)
            if err != nil {
                RestSetError(resp, BadInputError("JSON decode failed: %s " + err.Error()).Log())
                return
            }
        }
        info.BodyObj = bodyObj

        // Call the wrapped handler.
        sideEffects := NewRestSideEffects(mailer)
        respObj, restErr := handler(info, sideEffects)
        if restErr != nil {
            // Send the error response
            RestSetError(resp, restErr)
            return
        }

        // Marshall the success response
        jsonBytes, err := json.MarshalIndent(respObj, "", "    ")
        if err != nil {
            RestSetError(resp, InternalServerError("Error JSON-encoding Response").Log())
            return
        }
        resp.SetBody(map[string]interface{} {
            "http-body" : string(jsonBytes),
            "http-status" : http.StatusOK,
        })

        // Perform deferred side effects
        // This must occur after resp.SetBody
        sideEffects.Perform(req, resp)
    }
}
