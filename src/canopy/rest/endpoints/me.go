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
    "canopy/datalayer/cassandra_datalayer"
    "net/http"
)

func GET_me(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")

    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            writeNotLoggedInError(w);
            return
        }
    } else {
        writeNotLoggedInError(w);
        return
    }

    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    account, err := conn.LookupAccount(username_string)
    if err != nil {
        return
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\", \"username\" : \"%s\", \"email\" : \"%s\"}",
        account.Username(),
        account.Email())
    return
}

