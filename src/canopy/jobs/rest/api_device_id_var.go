/*
 * Copright 2014-2015 Canopy Services, Inc.
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
package rest

import (
    "canopy/datalayer"
    "time"
)

func GET__api__device__id__var(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    deviceIdString := info.URLVars["id"]
    sensorName := info.URLVars["var"]
    var device datalayer.Device
    var err error

    // TODO: fix anon devices
    if info.Account == nil {
        return nil, NotLoggedInError()
    }

    device, err = info.Account.Device(deviceIdString)
    if err != nil {
        // TODO: What errors to return here?
        return nil, InternalServerError("Device lookup failed")
    }

    doc := device.SDDLDocument()
    if doc == nil {
        return nil, URLNotFoundError()
    }

    varDef, err := doc.LookupVarDef(sensorName)
    if err != nil{
        return nil, URLNotFoundError()
    }

    samples, err := device.HistoricData(varDef, time.Now(), time.Now().Add(-59*time.Minute), time.Now())
    if err != nil {
        return nil, InternalServerError("Could not obtain sample data: " + err.Error())
    }

    // Convert samples to JSON
    out := map[string]interface{}{}
    out["result"] = "ok"
    out["samples"] = []interface{}{}
    for _, sample := range samples {
        out["samples"] = append(out["samples"].([]interface{}), map[string]interface{}{
            "t" : sample.Timestamp.Format(time.RFC3339),
            "v" : sample.Value,
        })
    }

    return out, nil
}
