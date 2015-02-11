// Copyright 2014 SimpleThings, Inc.
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
package adapter

import(
    "canopy/canolog"
    "canopy/config"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/mail"
    "canopy/pigeon"
    "canopy/rest/rest_errors"
    "encoding/base64"
    "errors"
    "encoding/json"
    "fmt"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "io/ioutil"
    "net/http"
    "strings"
)

type RestHandlerIn struct {
    Config config.Config
    CookieStore *sessions.CookieStore
    Mailer mail.MailClient
    PigeonSys *pigeon.PigeonSystem
}

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
    Session *sessions.Session
    Mailer mail.MailClient
    PigeonSys *pigeon.PigeonSystem
    URLVars map[string]string
}


type CanopyRestHandler func(http.ResponseWriter, *http.Request, CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError)

func basicAuthFromRequest(r *http.Request) (username string, password string, err error) {
    h, ok := r.Header["Authorization"]
    if !ok || len(h) == 0 {
        return "", "", errors.New("Authorization header not set")
    }
    parts := strings.SplitN(h[0], " ", 2)
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

func CanopyRestAdapter(fn CanopyRestHandler, in RestHandlerIn) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        info := CanopyRestInfo{
            Config: in.Config,
            Mailer: in.Mailer,
            PigeonSys: in.PigeonSys,
        }

        // Log request
        canolog.Info("Request: ", r.Method, r.URL, " BY ", r.RemoteAddr)

        // Get vars from URL if any
        info.URLVars = mux.Vars(r)

        // Connect to the database
        conn, err := cassandra_datalayer.NewDatalayerConnection(in.Config)
        if err != nil {
            rest_errors.NewDatabaseConnectionError().WriteTo(w)
            return
        }
        defer conn.Close()
        info.Conn = conn

        // Set standard headers
        w.Header().Set("Content-Type", "application/json")
        if (in.Config.OptAllowOrigin() != "") {
            w.Header().Set("Access-Control-Allow-Origin", in.Config.OptAllowOrigin())
            // Allow cross-origin cookies.
            // Client must also set "withCredentials" to ture on the
            // XMLHttpRequest.
            w.Header().Set("Access-Control-Allow-Credentials", "true")
        }

        // Check for BASIC AUTH
        username_string, password, err := basicAuthFromRequest(r)
        if err == nil {
            // was a UUID provided?
            if len(username_string) == 36 {
                device, err := conn.LookupDeviceByStringID(username_string)
                if err != nil {
                    w.WriteHeader(http.StatusUnauthorized)
                    fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
                    return
                }
                
                if device.SecretKey() != password {
                    w.WriteHeader(http.StatusUnauthorized)
                    fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
                    return
                }

                info.AuthType = CANOPY_REST_AUTH_DEVICE_BASIC
                info.Device = device

                // update last_seen for this device
                canolog.Info("Updating last seen")
                err = device.UpdateLastActivityTime(nil)
                if err != nil {
                    rest_errors.NewInternalServerError("Updating last seen time").WriteTo(w)
                    return
                }
                canolog.Info("Device BASIC auth provided")
            } else {
                // otherwise, assume user account username/password provided
                acct, err := conn.LookupAccountVerifyPassword(username_string, password)
                if err != nil {
                    if err == datalayer.InvalidPasswordError {
                        w.WriteHeader(http.StatusUnauthorized)
                        fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
                        return
                    } else {
                        w.WriteHeader(http.StatusInternalServerError);
                        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
                        return
                    }
                }
                
                canolog.Info("Basic auth provided")
                info.AuthType = CANOPY_REST_AUTH_BASIC
                info.Account = acct
            }
        }

        // Check for session-based AUTH
        session, _ := in.CookieStore.Get(r, "canopy-login-session")
        info.Session = session

        username, ok := session.Values["logged_in_username"]
        if ok {
            username_string, ok = username.(string)
            if ok && username_string != "" {
                canolog.Info("Looking up account: ", username_string)
                acct, err := conn.LookupAccount(username_string)
                if err != nil {
                    info.Session.Values["logged_in_username"] = ""
                    info.Session.Save(r, w)
                    w.WriteHeader(http.StatusInternalServerError);
                    fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
                    return
                }

                canolog.Info("Session auth provided")
                info.AuthType = CANOPY_REST_AUTH_SESSION
                info.Account = acct
            }
        }

        if info.Account == nil && info.Device == nil {
            canolog.Info("No auth provided")
        }
        // Parse the JSON payload
        // TODO: better way to figure out if there is a message body?
        var data map[string]interface{}
        bodyBytes, err := ioutil.ReadAll(r.Body)
        if err != nil {
            fmt.Fprintf(w, "{\"error\" : \"reading_body\"}")
            return
        }
        bodyString := string(bodyBytes)
        if bodyString != "" {
            decoder := json.NewDecoder(strings.NewReader(bodyString))
            err := decoder.Decode(&data)
            if err != nil {
                fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
                return
            }
            info.BodyObj = data
        }

        // Call the wrapped function.
        // The wrapped function may either write the response itself to <w>
        // (and return (nil, nil)), or it can return a JSON object that will be
        // marshalled by this wrapper, or it can return a CanopyRestError
        // object.
        jsonObj, restErr := fn(w, r, info)

        // Return the appropriate error, if an error occurred
        if restErr != nil {
            restErr.WriteTo(w)
            return
        }

        // On success, if jsonObj was returned:
        if jsonObj != nil {
            jsonBytes, err := json.MarshalIndent(jsonObj, "", "    ")
            if err != nil {
                return
            }
            jsonString := string(jsonBytes)
            fmt.Fprint(w, jsonString)
        }
    }
}

