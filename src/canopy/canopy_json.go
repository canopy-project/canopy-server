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
    Value float64 `json:"v"`
}

type jsonSamples struct {
    Samples []jsonSample `json:"samples"`
}

func devicesToJson(devices []*datalayer.CassandraDevice) (string, error) {
    var out jsonDevices

    for _, device := range devices {
        outDeviceClass := device.SDDLClass()
        if outDeviceClass != nil {
            outDeviceClassJson := outDeviceClass.Json()

            // get most recent value of each sensor/control
            propValues := map[string]jsonSample{}
            for _, prop := range outDeviceClass.Properties() {
                sensor, ok := prop.(*sddl.Sensor)
                if ok {
                    sample, err := device.GetCurrentSensorData(prop.JustName())
                    if err != nil {
                        continue
                    }
                    propValues[sensor.Name()] = jsonSample{
                        sample.Timestamp.Format(time.RFC3339),
                        sample.Value,
                    }
                }
            }

            out.Devices = append(
                out.Devices, jsonDevicesItem{
                    device.GetId().String(), 
                    device.GetFriendlyName(),
                    IsDeviceConnected(device.GetId().String()),
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

func samplesToJson(samples []datalayer.SensorSample) (string, error) {
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
