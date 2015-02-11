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
    "canopy/cloudvar"
    "canopy/config"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/sddl"
    "time"
    "github.com/gocql/gocql"
    "net/http"
    "fmt"
)

type ServiceResponse struct {
    HttpCode int
    Err error
    Response string
    Device datalayer.Device
}


// Process communication payload from device (via websocket. or REST)
//  {
//      "device_id" : "9dfe2a00-efe2-45f9-a84c-8afc69caf4e7", 
//        "sddl" : {
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
//
//  <conn> is an optional datalayer connection.  If provided, it is used.
//  Otherwise, a datalayer connection is opened by this routine.
//
//  <device> is the device that sent the communication.  If nil, then either
//  <deviceId> or, as a last resort, the payload's "device_id" will be used.
//
//  <deviceId> is a string device ID of the device that sent the communication.
//  This is ignored if <device> is not nil.  If nil, then the payload's
//  "device_id" will be used.
//
//  <secretKey> is the device's secret key. A secret key is required if
//  <device> is nil.  Either the value of <secretKey> or, as a last resort, the
//  payload's "secret_key" field will be used.
//
//  <payload> is a string containing the JSON payload.
func ProcessDeviceComm(
        cfg config.Config,
        conn datalayer.Connection, 
        device datalayer.Device, 
        deviceIdString string,
        secretKey string,
        payload string) ServiceResponse {
    var err error
    var out ServiceResponse
    var ok bool

    canolog.Info("ProcessDeviceComm STARTED")
    // If conn is nil, open a datalayer connection.
    if conn == nil {
        conn, err := cassandra_datalayer.NewDatalayerConnection(cfg)
        if err != nil {
            return ServiceResponse{
                HttpCode: http.StatusInternalServerError,
                Err: fmt.Errorf("Could not connect to database: %s", err),
                Response: `{"result" : "error", "error_type" : "could_not_connect_to_database"}`,
                Device: nil,
            }
        }
        defer conn.Close()
    }

    // Parse JSON payload
    var payloadObj map[string]interface{}
    err = json.Unmarshal([]byte(payload), &payloadObj)
    if err != nil{
        return ServiceResponse{
            HttpCode: http.StatusBadRequest,
            Err: fmt.Errorf("Error JSON decoding payload: %s", err),
            Response: `{"result" : "error", "error_type" : "decoding_paylaod"}`,
            Device: nil,
        }
    }

    // Device can be provided to this routine in one of three ways:
    // 1) <device> parameter
    // 2) <deviceId> parameter
    // 3) "device_id" field in payload
    if device == nil && deviceIdString != "" {
        // Parse UUID
        uuid, err := gocql.ParseUUID(deviceIdString)
        if err != nil {
            return ServiceResponse{
                HttpCode: http.StatusBadRequest,
                Err: fmt.Errorf("Invalid UUID %s: %s", deviceIdString, err),
                Response: `{"result" : "error", "error_type" : "device_uuid_required"}`,
                Device: nil,
            }
        }

        // Get secret key from payload if necessary
        if secretKey == "" {
            secretKey, ok = payloadObj["secret_key"].(string)
            if !ok {
                return ServiceResponse{
                    HttpCode: http.StatusBadRequest,
                    Err: fmt.Errorf("\"secret_key\" field must be string"),
                    Response: `{"result" : "error", "error_type" : "bad_payload"}`,
                    Device: nil,
                }
            }
        }

        // lookup device
        device, err = conn.LookupDeviceVerifySecretKey(uuid, secretKey)
        if err != nil {
            return ServiceResponse{
                HttpCode: http.StatusInternalServerError,
                Err: fmt.Errorf("Error looking up or verifying device: %s", err),
                Response: `{"result" : "error", "error_type" : "database_error"}`,
                Device: nil,
            }
        }
    }

    // Is "device_id" provided in payload?
    _, ok = payloadObj["device_id"]
    if ok {
        deviceIdStringFromPayload, ok := payloadObj["device_id"].(string)
        if !ok {
            return ServiceResponse{
                HttpCode: http.StatusBadRequest,
                Err: fmt.Errorf("\"device_id\" field must be string"),
                Response: `{"result" : "error", "error_type" : "bad_payload"}`,
                Device: nil,
            }
        }

        // Parse UUID
        uuid, err := gocql.ParseUUID(deviceIdStringFromPayload)
        if err != nil {
            return ServiceResponse{
                HttpCode: http.StatusBadRequest,
                Err: fmt.Errorf("Invalid UUID %s: %s", deviceIdStringFromPayload, err),
                Response: `{"result" : "error", "error_type" : "device_uuid_required"}`,
                Device: nil,
            }
        }

        // Is <device> already set?
        // If not: set it.
        // If so: ensure consistency
        if device == nil {
        
            // Get secret key from payload if necessary
            if secretKey == "" {
                secretKey, ok = payloadObj["secret_key"].(string)
                if !ok {
                    return ServiceResponse{
                        HttpCode: http.StatusBadRequest,
                        Err: fmt.Errorf("\"secret_key\" field must be string"),
                        Response: `{"result" : "error", "error_type" : "bad_payload"}`,
                        Device: nil,
                    }
                }
            }

            // Lookup device
            device, err = conn.LookupDeviceVerifySecretKey(uuid, secretKey)
            if err != nil {
                return ServiceResponse{
                    HttpCode: http.StatusInternalServerError,
                    Err: fmt.Errorf("Error looking up or verifying device: %s", err),
                    Response: `{"result" : "error", "error_type" : "database_error"}`,
                    Device: nil,
                }
            }
        } else {
            if device.ID().String() != deviceIdStringFromPayload {
                return ServiceResponse{
                    HttpCode: http.StatusBadRequest,
                    Err: fmt.Errorf("Inconsistent device ID: %s %s", device.ID().String(), deviceIdStringFromPayload),
                    Response: `{"result" : "error", "error_type" : "bad_payload"}`,
                    Device: nil,
                }
            }
        }
    }

    // If device wasn't provided at all, throw error.
    if device == nil {
        return ServiceResponse{
            HttpCode: http.StatusBadRequest,
            Err: fmt.Errorf("Device ID expected"),
            Response: `{"result" : "error", "error_type" : "bad_payload"}`,
            Device: nil,
        }
    }
    out.Device = device

    device.UpdateLastActivityTime(nil)

    // If "sddl" is present, create new / reconfigure Cloud Variables.
    _, ok = payloadObj["sddl"]
    if ok {
        updateMap, ok := payloadObj["sddl"].(map[string]interface{})
        if !ok {
            return ServiceResponse{
                HttpCode: http.StatusBadRequest,
                Err: fmt.Errorf("Expected object for \"sdd\" field"),
                Response: `{"result" : "error", "error_type" : "bad_payload"}`,
                Device: nil,
            }
        }
        err = device.ExtendSDDL(updateMap)
        if err != nil {
            return ServiceResponse{
                HttpCode: http.StatusInternalServerError,
                Err: fmt.Errorf("Error updating device's SDDL: %s", err),
                Response: `{"result" : "error", "error_type" : "database_error"}`,
                Device: nil,
            }
        }
    }

    // If "vars" is present, update value of all Cloud Variables (creating new
    // Cloud Variables as necessary)
    doc := device.SDDLDocument()
    _, ok = payloadObj["vars"]
    canolog.Info("vars present:", ok)
    if ok {
        varsMap, ok := payloadObj["vars"].(map[string]interface{})
        if !ok {
            return ServiceResponse{
                HttpCode: http.StatusBadRequest,
                Err: fmt.Errorf("Expected object for \"vars\" field"),
                Response: `{"result" : "error", "error_type" : "bad_payload"}`,
                Device: nil,
            }
        }
        canolog.Info("varsMap: ", varsMap)
        for varName, value := range varsMap {
            varDef, err := doc.LookupVarDef(varName)
            // TODO: an error doesn't necessarily mean prop should be created?
            canolog.Info("Looking up property ", varName)
            if (varDef == nil) {
                // Property doesn't exist.  Add it.
                canolog.Info("Not found.  Add property ", varName)
                // TODO: What datatype?
                // TODO: What other parameters?
                varDef, err = doc.AddVarDef(varName, sddl.DATATYPE_FLOAT32)
                if err != nil {
                    return ServiceResponse{
                        HttpCode: http.StatusInternalServerError,
                        Err: fmt.Errorf("Error creating cloud variable %s: %s", varName, err),
                        Response: `{"result" : "error", "error_type" : "database_error"}`,
                        Device: nil,
                    }
                }

                // save modified SDDL 
                // TODO: Save at the end?
                canolog.Info("SetSDDLDocument ", doc)
                err = device.SetSDDLDocument(doc)
                if err != nil {
                    return ServiceResponse{
                        HttpCode: http.StatusInternalServerError,
                        Err: fmt.Errorf("Error updating SDDL: %s", err),
                        Response: `{"result" : "error", "error_type" : "database_error"}`,
                        Device: nil,
                    }
                }
            }

            // Store property value.
            // Convert value datatype
            varVal, err := cloudvar.JsonToCloudVarValue(varDef, value)
            if err != nil {
                return ServiceResponse{
                    HttpCode: http.StatusInternalServerError,
                    Err: fmt.Errorf("Error converting JSON to propertyValue: %s", err),
                    Response: `{"result" : "error", "error_type" : "bad_payload"}`,
                    Device: nil,
                }
            }
            canolog.Info("InsertStample")
            err = device.InsertSample(varDef, time.Now(), varVal)
            if (err != nil) {
                return ServiceResponse{
                    HttpCode: http.StatusInternalServerError,
                    Err: fmt.Errorf("Error inserting sample %s: %s", varName, err),
                    Response: `{"result" : "error", "error_type" : "database_error"}`,
                    Device: nil,
                }
            }
        }
    }

    return ServiceResponse{
        HttpCode: http.StatusOK,
        Err: nil,
        Response: `{"result" : "ok"}`,
        Device: device,
    }
}
