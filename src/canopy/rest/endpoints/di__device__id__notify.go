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

//
// The device POSTs to this endpoint when a firmware routine such as
// "canopy_notify" is called.
//
//  POST /di/device/<UUID>
//  {
//      "notify-type" : "email",
//      "notify-msg" : <msg>
//  }
//
//  Effects:
//
//

import (
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/canolog"
    "canopy/notify"
    "encoding/json"
    "fmt"
    "github.com/gocql/gocql"
    "github.com/gorilla/mux"
    "net/http"
)

func POST_di__device__id__notify(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    deviceIdString := vars["id"]
    canolog.Info("/di/device/", deviceIdString, "/notify requested.")

    // Parse input as json
    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    // Get msg
    msgItf, ok := data["notify-msg"]
    if !ok {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Expected notify-msg\"}")
        return
    }
    msg, ok := msgItf.(string)
    if !ok {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Expected string for notify-msg value\"}")
        return
    }

    // Connect to database
    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    // Parse UUID
    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device UUID expected\"}");
        return
    }

    // Does device exist?  If not, create an anonymous device.
    device, err := conn.LookupOrCreateDevice(uuid, datalayer.ReadWriteAccess)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Error reading database\"}");
        return
    }

    err = notify.ProcessNotification(device, "email", msg);
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Error sending notification: ", err, "\"}");
        return
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\"}");
}
