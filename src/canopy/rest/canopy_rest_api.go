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
    "canopy/rest/adapter"
    "canopy/rest/endpoints"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
)

func AddRoutes(r *mux.Router, cfg config.Config) {
    store := sessions.NewCookieStore([]byte(cfg.OptProductionSecret()))

    // TODO: Need to handle allow-origin correctly!
    r.HandleFunc("/api/info", adapter.CanopyRestAdapter(endpoints.GET_info, cfg, store)).Methods("GET")
    r.HandleFunc("/api/create_account", adapter.CanopyRestAdapter(endpoints.POST_create_account, cfg, store)).Methods("POST")
    r.HandleFunc("/api/create_devices", adapter.CanopyRestAdapter(endpoints.POST_create_devices, cfg, store)).Methods("POST")
    r.HandleFunc("/api/device/{id}", adapter.CanopyRestAdapter(endpoints.GET_device__id, cfg, store)).Methods("GET")
    r.HandleFunc("/api/device/{id}", adapter.CanopyRestAdapter(endpoints.POST_device__id, cfg, store)).Methods("POST")
    r.HandleFunc("/api/device/{id}/{sensor}", adapter.CanopyRestAdapter(endpoints.GET_device__id__sensor, cfg, store)).Methods("GET")
    r.HandleFunc("/api/devices", adapter.CanopyRestAdapter(endpoints.GET_devices, cfg, store)).Methods("GET")
    r.HandleFunc("/api/me/devices", adapter.CanopyRestAdapter(endpoints.GET_devices, cfg, store)).Methods("GET")
    r.HandleFunc("/api/share", adapter.CanopyRestAdapter(endpoints.POST_share, cfg, store)).Methods("POST")
    r.HandleFunc("/api/finish_share_transaction", adapter.CanopyRestAdapter(endpoints.POST_finish_share_transaction, cfg, store)).Methods("POST")
    r.HandleFunc("/api/login", adapter.CanopyRestAdapter(endpoints.POST_login, cfg, store)).Methods("POST")
    r.HandleFunc("/api/logout", adapter.CanopyRestAdapter(endpoints.GET_POST_logout, cfg, store))
    r.HandleFunc("/api/me", adapter.CanopyRestAdapter(endpoints.GET_me, cfg, store)).Methods("GET")
    r.HandleFunc("/di/device/{id}", endpoints.POST_di__device__id).Methods("POST")
    r.HandleFunc("/di/device/{id}/notify", endpoints.POST_di__device__id__notify).Methods("POST")
}

