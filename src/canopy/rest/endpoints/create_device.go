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

func auth_user(conn datalayer.Connection, w http.ResponseWriter, r *http.Request) (datalayer.Account, error) {
    // First try BASIC AUTH
    username_string, password, err := basicAuthFromRequest(r)
    if err == nil {
        acct, err := conn.LookupAccountVerifyPassword(username_string, password)
        if err != nil {
            if err == datalayer.InvalidPasswordError {
                w.WriteHeader(http.StatusUnauthorized)
                fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
                return nil, err;
            } else {
                w.WriteHeader(http.StatusInternalServerError);
                fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
                return nil, err
            }
        }
        
        return acct, nil
    }

    // Next try session
    session, _ := store.Get(r, "canopy-login-session")

    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            writeNotLoggedInError(w);
            return nil, fmt.Errorf("Could not get username from session");
        }
    } else {
        writeNotLoggedInError(w);
        return nil, fmt.Errorf("Could not get username from session");
    }

    account, err := conn.LookupAccount(username_string)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return nil, err
    }

    return account, nil
}


func POST_create_device(w http.ResponseWriter, r *http.Request) {
    // TODO: Need to handle allow-origin correctly!
    writeStandardHeaders(w);

    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    acct, err := auth_user(conn, w, r)
    if err != nil {
        return
    }

    device, err := conn.CreateDevice("Pending Device", nil, "", datalayer.NoAccess);
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

    fmt.Fprintf(w, "{\"result\" : \"ok\", \"device_id\" : \"%s\"}", device.ID().String())
    return
}
