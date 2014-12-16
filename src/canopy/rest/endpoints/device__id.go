// Copyright 2014 SimpleThings, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package endpoints

import (
    "canopy/pigeon"
    "canopy/rest/adapter"
    "canopy/datalayer"
    "canopy/rest/rest_errors"
    "github.com/gocql/gocql"
    "net/http"
    "time"
)

func GET_device__id(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    // Used for anonymous devices
    deviceIdString := info.URLVars["id"]

    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        return nil, rest_errors.NewURLNotFoundError()
    }

    device, err := info.Conn.LookupDevice(uuid)
    if err != nil {
        // TODO: What errors to return here?
        return nil, rest_errors.NewInternalServerError("Device lookup failed")
    }
    out, err := deviceToJsonObj(device)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        return nil, rest_errors.NewInternalServerError("Generating JSON")
    }

    return out, nil
}

func POST_device__id(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    deviceIdString := info.URLVars["id"]

    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        return nil, rest_errors.NewURLNotFoundError()
    }

    var device datalayer.Device

    if info.Account != nil {
        device, err = info.Account.Device(uuid)
        if err != nil {
            // TODO: What errors to return here?
            return nil, rest_errors.NewInternalServerError("Device lookup failed")
        }
    } else if info.Device != nil {
        if deviceIdString != string(info.Device.IDString()) {
            // TODO: what error to return?
            return nil, rest_errors.NewInternalServerError("Device mismatch")
        }
        device = info.Device
    } else {
        return nil, rest_errors.NewNotLoggedInError()
    }

    /* Store cloud variable value.  */
    for fieldName, value := range info.BodyObj {
        if (fieldName == "__friendly_name") {
            friendlyName, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetName(friendlyName);
        } else if (fieldName == "__location_note") {
            locationNote, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetLocationNote(locationNote);
        } else {
            varDef, err := device.LookupVarDef(fieldName)
            if err != nil {
                /* TODO: Report warning in response*/
                continue;
            }
            varVal, err := JsonToCloudVarValue(varDef, value)
            if err != nil {
                /* TODO: Report warning in response*/
                continue;
            }
            device.InsertSample(varDef, time.Now(), varVal);
        }
    }

    msg := &pigeon.PigeonMessage {
        Data : info.BodyObj,
    }
    err = gPigeon.SendMessage(deviceIdString, msg, time.Duration(100*time.Millisecond))
    if err != nil {
        // TODO: Are there certain errors here that shouldn't be ignored?
        //return nil, rest_errors.NewInternalServerError("SendMessage failed")
    }

    return map[string]interface{} {
        "result" : "ok",
    }, nil
}
