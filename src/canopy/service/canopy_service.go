/*
 * Copyright 2014 SimpleThings Inc.
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
package service
import (
    "encoding/json"
    "canopy/canolog"
    "canopy/datalayer"
)

// Process communication payload from device (via websocket.  or REST?)
//  {
//      "device_id" : "9dfe2a00-efe2-45f9-a84c-8afc69caf4e7", 
//      "__sddl_update" : {
//          "control onoff" : { 
//              "datatype" : "float32" } 
//          } 
//      }
//  }
func ProcessDeviceComm(conn datalayer.Connection, payload string) string {
    var payloadObj map[string]interface{}
    var device datalayer.Device
    var deviceIdString string

    // parse payload
    err := json.Unmarshal([]byte(payload), &payloadObj)
    if err != nil{
        canolog.Error("Error JSON decoding payload: ", payload, err);
        return "";
    }

    // lookup device
    _, ok := payloadObj["device_id"]
    if ok {
        deviceIdString, ok = payloadObj["device_id"].(string)
        if !ok {
            canolog.Error("Expected string for device_id: ", payload);
            return "";
        }

        device, err = conn.LookupDeviceByStringID(deviceIdString)
        if err != nil {
            canolog.Error("Device not found: ", deviceIdString, err);
            return "";
        }
    } else {
        canolog.Error("device_id field mandatory: ", payload);
        return "";
    }

    // update SDDL if necessary
    _, ok = payloadObj["__sddl_update"] // TODO: call this "__sddl_extend" ?
    if ok {
        updateMap, ok := payloadObj["__sddl_update"].(map[string]interface{})
        if !ok {
            canolog.Error("Expected object for __sddl_update value");
            return "";
        }
        err = device.ExtendSDDLClass(updateMap)
    }
    return ""
}
