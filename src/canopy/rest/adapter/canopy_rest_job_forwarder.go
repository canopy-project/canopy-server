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

package adapter

import(
    "canopy/canolog"
    "canopy/jobqueue"
    "fmt"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "io/ioutil"
    "net/http"
    "runtime"
)

// This handler forwards an HTTP request along as a Pigeon job.
func CanopyRestJobForwarder(
        jobKey string, 
        cookieStore *sessions.CookieStore,
        allowOrigin string,
        pigeonClient jobqueue.Client) http.HandlerFunc {

    return func(w http.ResponseWriter, r *http.Request) {

        // Log crashes
        defer func() {
            r := recover()
            if r != nil {
                var buf [4096]byte
                runtime.Stack(buf[:], false)
                canolog.Error("PANIC ", r, string(buf[:]))
                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprintf(w, "{\"result\" : \"error\", \"error_type\" : \"crash\"}")
            }
        }()

        // Log request
        canolog.Info("Request: ", r.Method, r.URL, " BY ", r.RemoteAddr)

        // Check for session-based AUTH
        cookieUsername := ""
        if cookieStore != nil {
            session, _ := cookieStore.Get(r, "canopy-login-session")
            cookieUsername, _ = session.Values["logged_in_username"].(string)
        }

        // Read message body
        bodyBytes, err := ioutil.ReadAll(r.Body)
        if err != nil {
            fmt.Fprintf(w, "{\"error\" : \"reading_body\"}")
            return
        }
        bodyString := string(bodyBytes)

        // Launch backend job
        payload := map[string]interface{}{
            "url-vars" : mux.Vars(r),
            "auth-header" : r.Header["Authorization"],
            "cookie-username" : cookieUsername,
            "http-body" : bodyString,
        }
        //
        canolog.Info("Launching job", jobKey)
        respChan, err := pigeonClient.Launch(jobKey, payload)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(w, "{\"result\" : \"error\", \"error_type\" : \"failed_to_launch_job\"}")
            return
        }

        w.Header().Set("Content-Type", "application/json")
        if allowOrigin != "" {
            w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
        }

        // Wait for pigeon response
        resp := (<-respChan).Body()

        // Parse pigeon response
        httpStatus, ok := resp["http-status"].(int)
        if !ok {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(w, "{\"result\" : \"error\", \"error\" : \"Expected int http-status\"}")
            return
        }

        clearCookies, ok := resp["clear-cookies"].([]string)
        if ok {
            session, _ := cookieStore.Get(r, "canopy-login-session")
            for _, cookie := range clearCookies {
                canolog.Info("Clearing cookie: ", session, session.Values, cookie)
                session.Values[cookie] = ""
                canolog.Info("Cleared")
            }
            err := session.Save(r, w)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprintf(w, "{\"result\" : \"error\", \"error\" : \"error_saving_session\"}")
                return
            }
        }

        setCookies, ok := resp["set-cookies"].(map[string]string)
        if ok {
            session, _ := cookieStore.Get(r, "canopy-login-session")
            for key, value := range setCookies {
                canolog.Info("Setting cookie: ", key, ":", value)
                session.Values[key] = value
            }
            err := session.Save(r, w)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprintf(w, "{\"result\" : \"error\", \"error\" : \"error_saving_session\"}")
                return
            }
        }

        // Write HTTP Response
        w.WriteHeader(httpStatus)
        fmt.Fprint(w, resp["http-body"])
    }
}
