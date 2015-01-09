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
    "canopy/rest/adapter"
    "canopy/rest/rest_errors"
    "fmt"
    "github.com/gocql/gocql"
    "net/http"
    "time"
)

func GET_device__id__sensor(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    deviceIdString := info.URLVars["id"]
    sensorName := info.URLVars["sensor"]
    authorized := false;
    var device datalayer.Device

    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        return nil, rest_errors.NewURLNotFoundError()
    }

    //if info.Config.OptAllowAnonDevices() && device.PublicAccessLevel() > datalayer.NoAccess {
        canolog.Info("C");
        device, err = info.Conn.LookupDevice(uuid)
        if err != nil {
            // TODO: What errors to return here?
            return nil, rest_errors.NewInternalServerError("Device lookup failed")
        }
        authorized = true
    //} else {
    // TODO: fix anon devices
    canolog.Info("D");
    if info.Account == nil {
        return nil, rest_errors.NewNotLoggedInError()
    }

    device, err = info.Account.Device(uuid)
    if err != nil {
        // TODO: What errors to return here?
        return nil, rest_errors.NewInternalServerError("Device lookup failed")
    }

    authorized = true
    //}

    canolog.Info("E");
    if !authorized {
        // TODO: What is the correct error for this?
        return nil, rest_errors.NewURLNotFoundError()
    }

    canolog.Info("F");
    doc := device.SDDLDocument()
    if doc == nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device doesn't have any cloud variables\"}");
        return nil, nil
    }

    canolog.Info("G");
    varDef, err := doc.LookupVarDef(sensorName)
    if err != nil{
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device does not have cloud variable %s\"}", sensorName);
        return nil, nil
    }

    canolog.Info("H");
    samples, err := device.HistoricData(varDef, time.Now(), time.Now())
    if err != nil {
        fmt.Println(err)
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Could not obtain sample data\"}");
        return nil, nil
    }

    canolog.Info("I");
    out, err := samplesToJson(samples)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"generating_json\"} : ", err);
        return nil, nil
    }

    canolog.Info("J");
    fmt.Fprintf(w, out);
    return nil, nil
}
