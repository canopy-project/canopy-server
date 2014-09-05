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
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "encoding/json"
    "fmt"
    "net/http"
)

func POST_finish_share_transaction(w http.ResponseWriter, r *http.Request) {
    /*
     *  POST
     *  {
     *      "device_id" : <DEVICE_ID>,
     *  }
     *
     * TODO: Add to REST API documentation
     * TODO: Highly insecure!!!
     */
    var data map[string]interface{}
    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")

    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    deviceId, ok := data["device_id"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"device_id expected\"}")
        return
    }

    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            w.WriteHeader(http.StatusUnauthorized);
            fmt.Fprintf(w, "{\"error\" : \"not_logged_in\"");
            return
        }
    } else {
        w.WriteHeader(http.StatusUnauthorized);
        fmt.Fprintf(w, "{\"error\" : \"not_logged_in\"");
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
    if account == nil || err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return
    }

    device, err := conn.LookupDeviceByStringID(deviceId)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"device_lookup_failed\"}");
        return
    }

    /* Grant permissions to the user to access the device */
    err = device.SetAccountAccess(account, datalayer.ReadWriteAccess, datalayer.ShareRevokeAllowed)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"could_not_grant_access\"}");
        return
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\", \"device_friendly_name\" : \"%s\" }", device.Name());
    return
}
