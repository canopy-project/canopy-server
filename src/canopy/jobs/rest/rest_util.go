// Copright 2014-2015 Canopy Services, Inc.
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
package rest

import (
    "canopy/cloudvar"
    "canopy/datalayer"
    canotime "canopy/util/time"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strings"
    "time"
)

// TODO: Need to handle allow-origin correctly!
//var gConfAllowOrigin = "http://74.93.13.249:8080"

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

/*func writeStandardHeaders(w http.ResponseWriter) {
    w.Header().Set("Connection", "close")
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", gConfAllowOrigin)
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}*/

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

func deviceToJsonObj(device datalayer.Device, timestamp_type string) (map[string]interface{}, error) {
    statusJsonObj := map[string]interface{} {
        "ws_connected" : device.WSConnected(),
    }
    lastSeen := device.LastActivityTime()
    if lastSeen == nil {
        statusJsonObj["last_activity_time"] = nil
    } else {
        if timestamp_type == "epoch_us" {
            statusJsonObj["last_activity_time"] = canotime.EpochMicroseconds(*lastSeen)
        } else {
            statusJsonObj["last_activity_time"] = canotime.RFC3339(*lastSeen)
        }
    }

    out := map[string]interface{}{
        "device_id" : device.ID(),
        "friendly_name" : device.Name(),
        "location_note" : device.LocationNote(),
        "status" : statusJsonObj,
        "var_decls" : map[string]interface{} {},
        "secret_key" : device.SecretKey(),
        "vars" : map[string]interface{} {},
        "notifs" : []interface{} {},
    }

    sddlDoc := device.SDDLDocument()
    if sddlDoc != nil {
        jsn := sddlDoc.Json()
        if jsn != nil {
            out["var_decls"] = jsn
        }
    }

    outDoc := device.SDDLDocument()
    if outDoc != nil {
        // get most recent value of each sensor/control
        for _, varDef := range outDoc.VarDefs() {
            sample, err := device.LatestDataByName(varDef.Name())
            if err != nil {
                continue
            }
            if timestamp_type == "epoch_us" {
                out["vars"].(map[string]interface{})[varDef.Name()] = map[string]interface{} {
                    "t" : canotime.EpochMicroseconds(sample.Timestamp),
                    "v" : sample.Value,
                }
            } else {
                out["vars"].(map[string]interface{})[varDef.Name()] = map[string]interface{} {
                    "t" : canotime.RFC3339(sample.Timestamp),
                    "v" : sample.Value,
                }
            }
        }


        // Generate JSON for notifications
        //
        /*notifications, err := device.HistoricNotifications()
        canolog.Info("Reading notifications")
        if err != nil {
            canolog.Info("Error reading notifications %s", err)
            return nil, err
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
        }*/
    }
    
    return out, nil

}
func deviceToJsonString(device datalayer.Device, timestamp_type string) (string, error) {
    out, err := deviceToJsonObj(device, timestamp_type)
    if err != nil {
        return "", err;
    }

    jsn, err := json.Marshal(out)
    if err != nil {
        return "", err
    }
    return string(jsn), nil
}

func devicesToJsonObj(devices []datalayer.Device, timestamp_type string) (map[string]interface{}, error) {

    out := map[string]interface{} {
        "devices" : []interface{} {},
    }

    for _, device := range devices {
        deviceJsonObj, err := deviceToJsonObj(device, timestamp_type)
        if err != nil {
            continue
        }

        out["devices"] = append(out["devices"].([]interface{}), deviceJsonObj)
    }

    return out, nil
}

func devicesToJsonString(devices []datalayer.Device, timestamp_type string) (string, error) {
    out, err := devicesToJsonObj(devices, timestamp_type)
    if err != nil {
        return "", err;
    }

    jsn, err := json.Marshal(out)
    if err != nil {
        return "", err
    }
    return string(jsn), nil
}

func samplesToJsonObj(samples []cloudvar.CloudVarSample) (map[string]interface{}) {
    out := map[string]interface{}{}
    out["samples"] = []interface{}{}
    for _, sample := range samples {
        out["samples"] = append(out["samples"].([]interface{}), map[string]interface{}{
            "t" : canotime.EpochMicroseconds(sample.Timestamp),
            "v" : sample.Value,
        })
    }
    return out
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

