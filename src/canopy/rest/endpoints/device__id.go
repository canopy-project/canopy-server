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
    "canopy/datalayer/cassandra_datalayer"
    "canopy/pigeon"
    "encoding/json"
    "fmt"
    "github.com/gocql/gocql"
    "github.com/gorilla/mux"
    "net/http"
    "time"
)

func POST_device__id(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deviceIdString := vars["id"]
    //controlName := vars["control"]

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
        w.WriteHeader(http.StatusUnauthorized);
        fmt.Fprintf(w, "{\"error\" : \"not_logged_in2\"}");
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
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return
    }

    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device UUID expected\"}");
        return
    }

    device, err := account.Device(uuid)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Could not find or access device\"}");
        return
    }

    /* Parse input as json and just forward it along using pigeon */
    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err = decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    /* Store control value.  For now, use sensor_data table */
    for propName, value := range data {
        /* TODO: Verify that control is, in fact, a control according to SDDL
         * class */
        if (propName == "__friendly_name") {
            friendlyName, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetName(friendlyName);
        } else if (propName == "__location_note") {
            locationNote, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetLocationNote(locationNote);
        } else {
            prop, err := device.LookupProperty(propName)
            if err != nil {
                /* TODO: Report warning in response*/
                continue;
            }
            propVal, err := JsonToPropertyValue(prop, value)
            if err != nil {
                /* TODO: Report warning in response*/
                continue;
            }
            device.InsertSample(prop, time.Now(), propVal);
        }
    }

    msg := &pigeon.PigeonMessage {
        Data : data,
    }
    err = gPigeon.SendMessage(deviceIdString, msg, time.Duration(100*time.Millisecond))
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"SendMessage failed\"}");
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\"}");
    return
}
