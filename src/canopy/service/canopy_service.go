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
//        "var_config" : {
//          "optional inbound bool onoff" : {}
//        },
//        "vars" : {
//            "temperature" : 38.0f;
//            "gps" : {
//                "latitude" : 38.0f;
//                "longitude" : 38.0f;
//            }
//        }
//    }
//  }
func ProcessDeviceComm(conn datalayer.Connection, payload string) (datalayer.Device, string) {
    var payloadObj map[string]interface{}
    var device datalayer.Device
    var deviceIdString string

    // parse payload
    err := json.Unmarshal([]byte(payload), &payloadObj)
    if err != nil{
        canolog.Error("Error JSON decoding payload: ", payload, err);
        return nil, "";
    }

    // lookup device
    _, ok := payloadObj["device_id"]
    if ok {
        deviceIdString, ok = payloadObj["device_id"].(string)
        if !ok {
            canolog.Error("Expected string for device_id: ", payload);
            return nil, "";
        }

        device, err = conn.LookupDeviceByStringID(deviceIdString)
        if err != nil {
            canolog.Error("Device not found: ", deviceIdString, err);
            return nil, "";
        }
    } else {
        canolog.Error("device_id field mandatory: ", payload);
        return nil, "";
    }

    // update SDDL if necessary
    _, ok = payloadObj["var_config"]
    if ok {
        updateMap, ok := payloadObj["var_config"].(map[string]interface{})
        if !ok {
            canolog.Error("Expected object for var_config value");
            return nil, "";
        }
        err = device.ExtendSDDLClass(updateMap)
    }

    // store Cloud Variable values
    _, ok := payloadObj["vars"]
    if ok {
        vars, ok := payloadObj["vars"].(map[string]interface{})
        if !ok {
            canolog.Error("Expected object for vars value");
            return nil, "";
        }

        for k, v := range vars {
            sensor, err := sddlClass.LookupSensor(k)
            if err != nil {
                /* sensor not found */
                canolog.Warn("Unexpected key: ", k)
                continue
            }
            t := time.Now()
            // convert from JSON to Go
            v2, err := endpoints.JsonToPropertyValue(sensor, v)
            if err != nil {
                canolog.Warn(err)
                continue
            }
            // Insert converts from Go to Cassandra
            err = device.InsertSample(sensor, t, v2)
            if err != nil {
                canolog.Warn(err)
                continue
            }
        }
    }

    return device, ""
}
