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
// "canopy_easy_post_sample" is called.
//
//  POST /di/device/<UUID>
//  {
//      <sample_name> : <value>,
//      ...
//  }
//
//  Effects:
//
//  - If <UUID> does not correspond to a device, an "anonymous" device is
//    created.
//  - If <UUID> corresponds to a non-anonymous device, the request will fail
//    unless property authenticated.
//
//  - Any posted samples that do not correspond to an existing SDDL property,
//    that SDDL property is created.
//

import (
    "canopy/datalayer/cassandra_datalayer"
    "canopy/canolog"
    "encoding/json"
    "fmt"
    "github.com/gocql/gocql"
    "github.com/gorilla/mux"
    "net/http"
    "strings"
    "time"
)

func POST_di__device__id(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    deviceIdString := vars["id"]
    canolog.Info("/di/device/", deviceIdString, " requested.")

    // Parse input as json
    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
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
    device, err := conn.LookupOrCreateDevice(uuid)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Error reading database\"}");
        return
    }
    sddl := device.SDDLClass()

    // For each reported property, create SDDL property if necessary
    for propName, value := range data {
        if strings.HasPrefix(propName, "__") {
            continue;
        }
        prop := sddl.LookupPropertyOrNil(propName)
        if (prop == nil) {
            // Property doesn't exist.  Add it.
        }

        // Store property value.
        err = device.InsertSample(prop, time.Now(), value)
        if (err != nil) {
            canolog.Warn("Error inserting sample ", propName, ": ", err)
        }
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\"}");
}
