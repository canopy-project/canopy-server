/*
 * Copyright 2014 Gregory Prisament
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package rest

import (
    "canopy/config"
    "canopy/canolog"
    "canopy/jobqueue"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "net/http"
)

func rootRedirectHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/mgr/index.html", 301);
}

func AddRoutes(r *mux.Router, cfg config.Config, pigeonSys jobqueue.System) error {
    store := sessions.NewCookieStore([]byte(cfg.OptProductionSecret()))
    
    outbox := pigeonSys.NewOutbox()

    forwardAsPigeonJob := func(httpEndpoint, httpMethods, jobKey string) {
        canolog.Info("Registering route: ", httpEndpoint, "  to ", jobKey)
        r.HandleFunc(
            httpEndpoint, 
            CanopyRestJobForwarder(
                jobKey, 
                store, 
                cfg.OptAllowOrigin(), 
                outbox,
            ),
        ).Methods(httpMethods)
    }

    // TODO: Need to handle allow-origin correctly!
    // TODO: Can we automate all this?
    r.HandleFunc("/", rootRedirectHandler).Methods("GET")
    forwardAsPigeonJob("/api/activate", "POST", "api/activate")
    forwardAsPigeonJob("/api/create_devices", "POST", "api/create_devices")
    forwardAsPigeonJob("/api/create_user", "POST", "api/create_user")
    forwardAsPigeonJob("/api/device/{id}", "GET", "GET:api/device/id")
    forwardAsPigeonJob("/api/device/{id}", "POST", "POST:api/device/id")
    forwardAsPigeonJob("/api/device/{id}", "DELETE", "DELETE:api/device/id")
    forwardAsPigeonJob("/api/device/{id}/{var}", "GET", "api/device/id/var")
    forwardAsPigeonJob("/api/devices", "GET", "api/devices")
    forwardAsPigeonJob("/api/finish_share_transaction", "POST", "api/finish_share_transaction")
    forwardAsPigeonJob("/api/info", "GET", "api/info")
    forwardAsPigeonJob("/api/login", "POST", "api/login")
    forwardAsPigeonJob("/api/logout", "GET", "api/logout")
    forwardAsPigeonJob("/api/logout", "POST", "api/logout")
    forwardAsPigeonJob("/api/user/self", "GET", "GET:api/user/self")
    forwardAsPigeonJob("/api/user/self", "POST", "POST:api/user/self")
    forwardAsPigeonJob("/api/user/self", "DELETE", "DELETE:api/user/self")
    forwardAsPigeonJob("/api/me/devices", "GET", "api/devices")
    forwardAsPigeonJob("/api/reset_password", "POST", "api/reset_password")
    forwardAsPigeonJob("/api/share", "POST", "api/share")

    return nil
}
