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
    "canopy/mail"
    "canopy/rest/adapter"
    "canopy/rest/endpoints"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "net/http"
)

func rootRedirectHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/mgr/index.html", 301);
}

func AddRoutes(r *mux.Router, cfg config.Config) error {
    store := sessions.NewCookieStore([]byte(cfg.OptProductionSecret()))
    
    mailer, err := mail.NewMailClient(cfg)
    if err != nil {
        return err
    }

    extra := adapter.RestHandlerIn{
        Config: cfg,
        CookieStore: store,
        Mailer: mailer,
   }

    // TODO: Need to handle allow-origin correctly!
    r.HandleFunc("/", rootRedirectHandler).Methods("GET")
    r.HandleFunc("/api/activate", adapter.CanopyRestAdapter(endpoints.POST_activate, extra)).Methods("POST")
    r.HandleFunc("/api/info", adapter.CanopyRestAdapter(endpoints.GET_info, extra)).Methods("GET")
    r.HandleFunc("/api/create_account", adapter.CanopyRestAdapter(endpoints.POST_create_account, extra)).Methods("POST")
    r.HandleFunc("/api/create_devices", adapter.CanopyRestAdapter(endpoints.POST_create_devices, extra)).Methods("POST")
    r.HandleFunc("/api/device/{id}", adapter.CanopyRestAdapter(endpoints.GET_device__id, extra)).Methods("GET")
    r.HandleFunc("/api/device/{id}", adapter.CanopyRestAdapter(endpoints.POST_device__id, extra)).Methods("POST")
    r.HandleFunc("/api/device/{id}/{sensor}", adapter.CanopyRestAdapter(endpoints.GET_device__id__sensor, extra)).Methods("GET")
    r.HandleFunc("/api/devices", adapter.CanopyRestAdapter(endpoints.GET_devices, extra)).Methods("GET")
    r.HandleFunc("/api/me/devices", adapter.CanopyRestAdapter(endpoints.GET_devices, extra)).Methods("GET")
    r.HandleFunc("/api/share", adapter.CanopyRestAdapter(endpoints.POST_share, extra)).Methods("POST")
    r.HandleFunc("/api/finish_share_transaction", adapter.CanopyRestAdapter(endpoints.POST_finish_share_transaction, extra)).Methods("POST")
    r.HandleFunc("/api/login", adapter.CanopyRestAdapter(endpoints.POST_login, extra)).Methods("POST")
    r.HandleFunc("/api/logout", adapter.CanopyRestAdapter(endpoints.GET_POST_logout, extra))
    r.HandleFunc("/api/me", adapter.CanopyRestAdapter(endpoints.GET_me, extra)).Methods("GET")
    r.HandleFunc("/api/me", adapter.CanopyRestAdapter(endpoints.POST_me, extra)).Methods("POST")
    r.HandleFunc("/di/device/{id}", endpoints.POST_di__device__id).Methods("POST")
    r.HandleFunc("/di/device/{id}/notify", endpoints.POST_di__device__id__notify).Methods("POST")

    return nil
}

