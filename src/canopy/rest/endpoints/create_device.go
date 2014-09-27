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
package endpoints

import (
    "fmt"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "net/http"
)

func POST_create_device(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);

    username, password, err := basicAuthFromRequest(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        fmt.Fprintf(w, "{\"error\" : \"bad_credentials\"}")
        return
    }

    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    acct, err := conn.LookupAccountVerifyPassword(username, password)
    if err != nil {
        if err == datalayer.InvalidPasswordError {
            w.WriteHeader(http.StatusUnauthorized)
            fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
            return;
        } else {
            w.WriteHeader(http.StatusInternalServerError);
            fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
            return
        }
    }

    device, err := conn.CreateDevice("Pending Device", nil, datalayer.NoAccess);
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Println(err)
        fmt.Fprintf(w, "{\"error\" : \"device_creation_failed\"}");
        return
    }

    err = device.SetAccountAccess(acct, datalayer.ReadWriteAccess, datalayer.ShareRevokeAllowed);
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"could_not_grant_access\"}");
        return
    }

    fmt.Fprintf(w, "{\"success\" : true, \"device_id\" : \"%s\"}", device.ID().String())
    return
}
