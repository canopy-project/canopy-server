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
package main

import (
    "canopy/datalayer"
    "canopy/sddl"
    "encoding/json"
    "time"
)

type jsonDevices struct {
    Devices []jsonDevicesItem `json:"devices"`
}

type jsonDevicesItem struct {
    DeviceId string `json:"device_id"`
    FriendlyName string `json:"friendly_name"`
    Connected bool `json:"connected"`
    ClassItems map[string]interface{} `json:"sddl_class"`
    PropValues map[string]jsonSample `json:"property_values"`
}

type jsonSample struct {
    Time string `json:"t"`
    Value interface{} `json:"v"`
}

type jsonSamples struct {
    Samples []jsonSample `json:"samples"`
}

func devicesToJson(devices []datalayer.Device) (string, error) {

    out := jsonDevices{[]jsonDevicesItem{}};

    for _, device := range devices {
        outDeviceClass := device.SDDLClass()
        if outDeviceClass != nil {
            outDeviceClassJson := outDeviceClass.Json()

            // get most recent value of each sensor/control
            propValues := map[string]jsonSample{}
            for _, prop := range outDeviceClass.Properties() {
                sensor, ok := prop.(*sddl.Sensor)
                if ok {
                    sample, err := device.LatestDataByPropertyName(prop.Name())
                    if err != nil {
                        continue
                    }
                    propValues[sensor.Name()] = jsonSample{
                        sample.Timestamp.Format(time.RFC3339),
                        sample.Value,
                    }
                }
                control, ok := prop.(*sddl.Control)
                if ok {
                    sample, err := device.LatestDataByPropertyName(prop.Name())
                    if err != nil {
                        continue
                    }
                    propValues[control.Name()] = jsonSample{
                        sample.Timestamp.Format(time.RFC3339),
                        sample.Value,
                    }
                }

            }

            out.Devices = append(
                out.Devices, jsonDevicesItem{
                    device.ID().String(), 
                    device.Name(),
                    IsDeviceConnected(device.ID().String()),
                    outDeviceClassJson,
                    propValues})
        }
    }

    jsn, err := json.Marshal(out)
    if err != nil {
        return "", err
    }
    return string(jsn), nil
}

func samplesToJson(samples []sddl.PropertySample) (string, error) {
    out := jsonSamples{[]jsonSample{}}
    for _, sample := range samples {
        out.Samples = append(out.Samples, jsonSample{
            sample.Timestamp.Format(time.RFC3339),
            sample.Value})
    }

    jsn, err := json.Marshal(out)
    if err != nil {
        return "", err
    }
    return string(jsn), nil
}
