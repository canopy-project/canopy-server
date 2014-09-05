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
    "encoding/json"
    "fmt"
    "canopy/datalayer/cassandra_datalayer"
    "net/http"
)

func POST_create_account(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);

    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    username, ok := data["username"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"string_username_expected\"}")
        return
    }

    password, ok := data["password"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"string_password_expected\"}")
        return
    }

    session, _ := store.Get(r, "canopy-login-session")
    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()
    account, err := conn.LookupAccountVerifyPassword(username, password)
    if err == nil {
        session.Values["logged_in_username"] = username
        err := session.Save(r, w)
        if err != nil {
            fmt.Fprintf(w, "{\"error\" : \"saving_session\"}")
            return
        }
        fmt.Fprintf(w, "{\"result\" : \"ok\", \"username\" : \"%s\", \"email\" : \"%s\"}",
            account.Username(),
            account.Email())
        return
    } else {
        writeIncorrectUsernameOrPasswordError(w);
        return
    }
}

