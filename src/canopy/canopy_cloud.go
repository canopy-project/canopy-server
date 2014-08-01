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
    "fmt"
    "time"
    "encoding/json"
    "code.google.com/p/go.net/websocket"
    "io"
    "net"
    "canopy/datalayer"
    "canopy/pigeon"
    "canopy/sddl"
)

// Process JSON message from the client
func processPayload(dl *datalayer.CassandraDatalayer, payload string, cnt int32) string{
    var payloadObj map[string]interface{}
    var device *datalayer.CassandraDevice
    var deviceIdString string
    var sddlClass *sddl.Class

    err := json.Unmarshal([]byte(payload), &payloadObj)
    if err != nil{
        fmt.Println("Error JSON decoding payload: ", payload)
        return "";
    }

    /* Lookup device */
    _, ok := payloadObj["device_id"]
    if ok {
        deviceIdString, ok = payloadObj["device_id"].(string)
        if !ok {
            fmt.Println("Expected string for device_id")
            return "";
        }

        device, err = dl.LookupDeviceByStringId(deviceIdString)
        if err != nil {
            fmt.Println("Device not found: ", deviceIdString, err)
            return "";
        }
    } else {
            fmt.Println("device-id field mandatory")
            return "";
    }

    /* Store SDDL class */
    _, ok = payloadObj["sddl"]
    if ok {
        sddlJson, ok := payloadObj["sddl"].(map[string]interface{})
        if !ok {
            fmt.Println("Expected object for SDDL")
            return "";
        }
        sddlClass, err = sddl.ParseClass("anonymous", sddlJson)
        if err != nil {
            fmt.Println("Failed parsing sddl class definition: ", err)
            return "";
        }

        err = device.SetSDDLClass(sddlClass)
        if err != nil {
            fmt.Println("Error storing SDDL class during processPayload")
            return "";
        }
    } else {
            fmt.Println("sddl field mandatory")
            return "";
    }


    /* Store sensor data */
    if cnt % 100 != 0 {
        for k, v := range payloadObj {
            /* hack */
            if k == "device_id" || k == "sddl" {
                continue
            }
            sensor, _ := sddlClass.LookupSensor(k)
            if sensor != nil {
                err = nil
                t := time.Now();
                switch sensor.Datatype() {
                case sddl.DATATYPE_VOID:
                    err = device.InsertSensorSample_void(k, t)
                case sddl.DATATYPE_STRING:
                    value, ok := v.(string)
                    if (!ok) {
                        fmt.Println("Expected string value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_string(k, t, value)
                case sddl.DATATYPE_BOOL:
                    value, ok := v.(bool)
                    if (!ok) {
                        fmt.Println("Expected boolean value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_bool(k, t, value)
                case sddl.DATATYPE_INT8:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_int8(k, t, int8(value))
                case sddl.DATATYPE_UINT8:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_uint8(k, t, uint8(value))
                case sddl.DATATYPE_INT16:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_int16(k, t, int16(value))
                case sddl.DATATYPE_UINT16:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_uint16(k, t, uint16(value))
                case sddl.DATATYPE_INT32:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_int32(k, t, int32(value))
                case sddl.DATATYPE_UINT32:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_uint32(k, t, uint32(value))
                case sddl.DATATYPE_FLOAT32:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_float32(k, t, float32(value))
                case sddl.DATATYPE_FLOAT64:
                    value, ok := v.(float64)
                    if (!ok) {
                        fmt.Println("Expected numeric value for ", k)
                        return ""
                    }
                    err = device.InsertSensorSample_float64(k, t, value)
                case sddl.DATATYPE_DATETIME:
                    /*value, ok := v.(string)
                    if (!ok) {
                        fmt.Println("Expected string datatime value for ", k)
                        return ""
                    }
                    value_t : time.Time(value)
                    err = device.InsertSensorSample_datetime(k, t, value_t)*/
                    fmt.Println("Datetime properties not yet supported")
                }
            } else {
                /* sensor not found */
                fmt.Println("Unexpected key: ", k)
            }
        }
    }

    return deviceIdString;
}

func IsDeviceConnected(deviceIdString string) bool {
    return (gPigeon.Mailbox(deviceIdString) != nil)
}

// Main websocket server routine.
// This event loop runs until the websocket connection is broken.
func CanopyWebsocketServer(ws *websocket.Conn) {

    var mailbox *pigeon.PigeonMailbox
    var cnt int32
    
    cnt = 0

    // connect to cassandra
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()

    for {
        var in string

        // check for message from client
        ws.SetReadDeadline(time.Now().Add(100*time.Millisecond))
        err := websocket.Message.Receive(ws, &in)
        if err == nil {
            // success, payload received
            cnt++;
            deviceId := processPayload(dl, in, cnt)
            if deviceId != "" && mailbox == nil {
                mailbox = gPigeon.CreateMailbox(deviceId)
            }
        } else if err == io.EOF {
            // connection closed
            if mailbox != nil {
                mailbox.Close()
            }
            return;
        } else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
            // timeout reached, no data for me this time
        } else {
            fmt.Println("Unexpected error:", err);
        }

        if mailbox != nil {
            msg, _ := mailbox.RecieveMessage(time.Duration(100*time.Millisecond))
            if msg != nil {
                msgString, err := json.Marshal(msg)
                if err != nil {
                    fmt.Println("Unexpected error:", err);
                }
                
                websocket.Message.Send(ws, msgString)
            }
        }
    }
}
