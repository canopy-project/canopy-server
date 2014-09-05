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
    //r.HandleFunc("/create_account", createAccountHandler)
    //r.HandleFunc("/create_device", createDeviceHandler)
    /*r.HandleFunc("/device/{id}", getDeviceInfoHandler).Methods("GET");*/
    //r.HandleFunc("/device/{id}", controlHandler).Methods("POST");
    //r.HandleFunc("/device/{id}/{sensor}", sensorDataHandler).Methods("GET");
    //r.HandleFunc("/devices", devicesHandler)
    //r.HandleFunc("/share", shareHandler)
    //r.HandleFunc("/finish_share_transaction", finishShareTransactionHandler)
    r.HandleFunc("/new/login", endpoints.POST_login).Methods("POST");
    //r.HandleFunc("/logout", logoutHandler);
    //r.HandleFunc("/me", meHandler);
    return r
}

