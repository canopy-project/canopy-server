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
    "canopy/cloudvar"
    "canopy/datalayer"
    "canopy/pigeon"
    "canopy/sddl"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strings"
    "time"
)

var gConfAllowOrigin = ""
var gPigeon = pigeon.InitPigeonSystem()
func writeDatabaseConnectionError(w http.ResponseWriter) {
    w.WriteHeader(http.StatusInternalServerError);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "could_not_connect_to_database"}`);
}
func writeNotLoggedInError(w http.ResponseWriter) {
    w.WriteHeader(http.StatusUnauthorized);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "not_logged_in"}`);
}

func writeAccountLookupFailedError(w http.ResponseWriter) {
    w.WriteHeader(http.StatusInternalServerError);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "account_lookup_failed"}`);
}

func writeIncorrectUsernameOrPasswordError(w http.ResponseWriter) {
    w.WriteHeader(http.StatusUnauthorized);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "incorrect_username_or_password"}`);
}

func writeStandardHeaders(w http.ResponseWriter) {
    w.Header().Set("Connection", "close")
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", gConfAllowOrigin)
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func basicAuthFromRequest(r *http.Request) (username string, password string, err error) {
    h, ok := r.Header["Authorization"]
    if !ok || len(h) == 0 {
        return "", "", errors.New("Authorization header not set")
    }
    parts := strings.SplitN(h[0], " ", 2)
    if len(parts) != 2 {
        return "", "", errors.New("Authentication header malformed")
    }
    if parts[0] != "Basic" {
        return "", "", errors.New("Expected basic authentication")
    }
    encodedVal := parts[1]
    decodedVal, err := base64.StdEncoding.DecodeString(encodedVal)
    if err != nil {
        return "", "", errors.New("Authentication header malformed")
    }
    parts = strings.Split(string(decodedVal), ":")
    if len(parts) != 2 {
        return "", "", errors.New("Authentication header malformed")
    }
    return parts[0], parts[1], nil
}

// converts based on SDDL property datatype:
// SDDL dataype        JSON type(in)   Go type (out)
// ----------------------------------------------
// void                  nil     -->    nil
// string                string  -->    string
// bool                  bool    -->    bool
// int8                  float64 -->    int8
// uint8                 float64 -->    uint8
// int16                 float64 -->    int16
// uint16                float64 -->    uint16
// int32                 float64 -->    int32
// uint32                float64 -->    uint32
// float32               float64 -->    float32
// float64               float64 -->    float64
// datetime              string  -->    time.Time
//
func JsonToCloudVarValue(varDef sddl.VarDef, value interface{}) (interface{}, error) {
    switch varDef.Datatype() {
    case sddl.DATATYPE_VOID:
        return nil, nil
    case sddl.DATATYPE_STRING:
        v, ok := value.(string)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects string value for %s", varDef.Name())
        }
        return v, nil
    case sddl.DATATYPE_BOOL:
        v, ok := value.(bool)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects bool value for %s", varDef.Name())
        }
        return v, nil
    case sddl.DATATYPE_INT8:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return int8(v), nil
    case sddl.DATATYPE_UINT8:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return uint16(v), nil
    case sddl.DATATYPE_INT16:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return int16(v), nil
    case sddl.DATATYPE_UINT16:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return uint16(v), nil
    case sddl.DATATYPE_INT32:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return int32(v), nil
    case sddl.DATATYPE_UINT32:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return uint32(v), nil
    case sddl.DATATYPE_FLOAT32:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return float32(v), nil
    case sddl.DATATYPE_FLOAT64:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return v, nil
    case sddl.DATATYPE_DATETIME:
        v, ok := value.(string)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects string value for %s", varDef.Name())
        }
        tval, err := time.Parse(time.RFC3339, v)
        if err != nil {
            return nil, fmt.Errorf("JsonToCloudVarValue expects RFC3339 formatted time value for %s", varDef.Name())
        }
        return tval, nil
    default:
        return nil, fmt.Errorf("InsertSample unsupported datatype ", varDef.Datatype())
    }
}

type jsonDevices struct {
    Devices []jsonDevicesItem `json:"devices"`
}

type jsonDevicesItem struct {
    DeviceId string `json:"device_id"`
    FriendlyName string `json:"friendly_name"`
    Connected bool `json:"connected"`
    ClassItems map[string]interface{} `json:"sddl_class"`
    PropValues map[string]jsonSample `json:"property_values"`
    Notifications []jsonNotification `json:"notifications"`
}

type jsonSample struct {
    Time string `json:"t"`
    Value interface{} `json:"v"`
}

type jsonSamples struct {
    Samples []jsonSample `json:"samples"`
}

type jsonNotification struct {
    Time string `json:"t"`
    Dismissed bool `json:"dismissed"`
    Msg string `json:"msg"`
}

func deviceToJson(device datalayer.Device) (string, error) {
    // TODO: Unify this and devicesToJson
    out := jsonDevicesItem{}

    outDoc := device.SDDLDocument()
    if outDoc != nil {
        outDocJson := outDoc.Json()

        // get most recent value of each sensor/control
        varValues := map[string]jsonSample{}
        for _, varDef := range outDoc.VarDefs() {
            sample, err := device.LatestDataByName(varDef.Name())
            if err != nil {
                continue
            }
            varValues[varDef.Name()] = jsonSample{
                sample.Timestamp.Format(time.RFC3339),
                sample.Value,
            }
        }


        // Generate JSON for notifications
        //
        notifications, err := device.HistoricNotifications()
        canolog.Info("Reading notifications")
        if err != nil {
            canolog.Info("Error reading notifications %s", err)
            return "", err
        }

        outNotifications := []jsonNotification{};
        for _, notification := range notifications {
            outNotifications = append(
                    outNotifications, 
                    jsonNotification{
                        notification.Datetime().Format(time.RFC3339),
                        notification.IsDismissed(),
                        notification.Msg(),
                    })
        }

        out = jsonDevicesItem{
                device.ID().String(), 
                device.Name(),
                IsDeviceConnected(device.ID().String()),
                outDocJson,
                varValues,
                outNotifications,
        }
    }

    jsn, err := json.Marshal(out)
    if err != nil {
        return "", err
    }
    return string(jsn), nil
}

func devicesToJson(devices []datalayer.Device) (string, error) {

    out := jsonDevices{[]jsonDevicesItem{}};

    for _, device := range devices {
        outDoc := device.SDDLDocument()
        if outDoc != nil {
            outDocJson := outDoc.Json()

            // get most recent value of each cloud variable
            varValues := map[string]jsonSample{}
            for _, varDef := range outDoc.VarDefs() {
                sample, err := device.LatestDataByName(varDef.Name())
                if err != nil {
                    continue
                }
                varValues[varDef.Name()] = jsonSample{
                    sample.Timestamp.Format(time.RFC3339),
                    sample.Value,
                }
            }

            out.Devices = append(
                out.Devices, jsonDevicesItem{
                    device.ID().String(), 
                    device.Name(),
                    IsDeviceConnected(device.ID().String()),
                    outDocJson,
                    varValues,
                    nil})
        }
    }

    jsn, err := json.Marshal(out)
    if err != nil {
        return "", err
    }
    return string(jsn), nil
}

func samplesToJson(samples []cloudvar.CloudVarSample) (string, error) {
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

func IsDeviceConnected(deviceIdString string) bool {
    return (gPigeon.Mailbox(deviceIdString) != nil)
}
