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
    "canopy/canolog"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "fmt"
    "github.com/gocql/gocql"
    "github.com/gorilla/mux"
    "net/http"
    "time"
)

func GET_device__id__sensor(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deviceIdString := vars["id"]
    sensorName := vars["sensor"]
    authorized := false;
    writeStandardHeaders(w);

    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device UUID expected\"}");
        return
    }

    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    device, err := conn.LookupDevice(uuid)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"device_lookup_failed\"}");
        return
    }

    if device.PublicAccessLevel() > datalayer.NoAccess || true { // TODO FIX
        authorized = true
    } else {
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

        account, err := conn.LookupAccount(username_string)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError);
            fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
            return
        }

        device, err = account.Device(uuid)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest);
            fmt.Fprintf(w, "{\"error\" : \"Could not find or access device\"}");
            return
        }

        authorized = true
    }
    canolog.Info("D")

    if !authorized {
        w.WriteHeader(http.StatusUnauthorized);
        fmt.Fprintf(w, "{\"error\" : \"Not authorized to access sensor data\"}");
        return
    }

    sddlClass := device.SDDLClass()
    if sddlClass == nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device doesn't have any sensors\"}");
        return
    }

    property, err := sddlClass.LookupProperty(sensorName)
    if err != nil{
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device does not have property %s\"}", sensorName);
        return
    }

    samples, err := device.HistoricData(property, time.Now(), time.Now())
    if err != nil {
        fmt.Println(err)
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Could not obtain sample data\"}");
        return
    }

    out, err := samplesToJson(samples)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"generating_json\"} : ", err);
        return
    }

    fmt.Fprintf(w, out);
    return
}
