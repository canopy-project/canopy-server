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
    "canopy/rest/endpoints"
    "github.com/gorilla/mux"
    "net/http"
)

func GetRestHandler() http.Handler {
    r := mux.NewRouter()
    /* TODO: add "/api" to the front of each of these.
     * (Will need to update application.)
     * 
     * Doing this will minimize collisions when we start to allow /USERNAME/DEVICE
     */
    r.HandleFunc("/api/info", endpoints.GET_info).Methods("GET")
    r.HandleFunc("/create_account", endpoints.POST_create_account).Methods("POST")
    r.HandleFunc("/create_device", endpoints.POST_create_device).Methods("POST")
    r.HandleFunc("/device/{id}", endpoints.POST_device__id).Methods("POST")
    r.HandleFunc("/device/{id}/{sensor}", endpoints.GET_device__id__sensor).Methods("GET")
    r.HandleFunc("/devices", endpoints.GET_devices).Methods("GET")
    r.HandleFunc("/share", endpoints.POST_share).Methods("POST")
    r.HandleFunc("/finish_share_transaction", endpoints.POST_finish_share_transaction).Methods("POST")
    r.HandleFunc("/login", endpoints.POST_login).Methods("POST")
    r.HandleFunc("/logout", endpoints.GET_POST_logout)
    r.HandleFunc("/me", endpoints.GET_me)
    return r
}

