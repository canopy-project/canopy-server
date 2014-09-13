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
package webapp

import (
    "fmt"
    "github.com/gorilla/mux"
    "net/http"
)

func AddRoutes(r *mux.Router) {
    r.HandleFunc("/device/{id}", GET_device__id).Methods("GET")
}

func GET_device__id(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    uuidString := vars["id"]
    fmt.Fprint(w, `<html>
<body>
Your UUID is <b>`, uuidString ,`</b>
</body>
</html>`)
}
