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
    "time"
    "encoding/json"
    "code.google.com/p/go.net/websocket"
    "io"
    "net"
    "canopy/canolog"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/pigeon"
    "canopy/sddl"
)

// Process JSON message from the client
func processPayload(conn datalayer.Connection, payload string, cnt int32) string{
    var payloadObj map[string]interface{}
    var device datalayer.Device
    var deviceIdString string
    var sddlClass *sddl.Class

    err := json.Unmarshal([]byte(payload), &payloadObj)
    if err != nil{
        canolog.Error("Error JSON decoding payload: ", payload, err);
        return "";
    }

    /* Lookup device */
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

    /* Store SDDL class */
    _, ok = payloadObj["sddl"]
    if ok {
        sddlJson, ok := payloadObj["sddl"].(map[string]interface{})
        if !ok {
            canolog.Error("Expected object for SDDL");
            return "";
        }
        sddlClass, err = sddl.ParseClass("anonymous", sddlJson)
        if err != nil {
            canolog.Error("Failed parsing sddl class definition: ", err);
            return "";
        }

        err = device.SetSDDLClass(sddlClass)
        if err != nil {
            canolog.Error("Error storing SDDL class during processPayload")
            return "";
        }
    } else {
            canolog.Error("sddl field mandatory:", payload)
            return "";
    }


    /* Store sensor data */
    if cnt % 100 != 0 {
        for k, v := range payloadObj {
            /* hack */
            if k == "device_id" || k == "sddl" {
                continue
            }
            sensor, err := sddlClass.LookupSensor(k)
            if err != nil {
                /* sensor not found */
                canolog.Warn("Unexpected key: ", k)
                continue
            }
            t := time.Now()
            // convert from JSON to Go
            v2, err := jsonToPropertyValue(sensor, v)
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

    return deviceIdString;
}

func IsDeviceConnected(deviceIdString string) bool {
    return (gPigeon.Mailbox(deviceIdString) != nil)
}

// Main websocket server routine.
// This event loop runs until the websocket connection is broken.
func CanopyWebsocketServer(ws *websocket.Conn) {

    canolog.Websocket("Websocket connection established")

    var mailbox *pigeon.PigeonMailbox
    var cnt int32
    
    cnt = 0

    // connect to cassandra
    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        canolog.Error("Could not connect to database: ", err)
        return
    }
    defer conn.Close()

    for {
        var in string

        // check for message from client
        ws.SetReadDeadline(time.Now().Add(100*time.Millisecond))
        err := websocket.Message.Receive(ws, &in)
        if err == nil {
            // success, payload received
            cnt++;
            deviceId := processPayload(conn, in, cnt)
            if deviceId != "" && mailbox == nil {
                mailbox = gPigeon.CreateMailbox(deviceId)
            }
        } else if err == io.EOF {
            canolog.Websocket("Websocket connection closed")
            // connection closed
            if mailbox != nil {
                mailbox.Close()
            }
            return;
        } else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
            // timeout reached, no data for me this time
        } else {
            canolog.Error("Unexpected error: ", err)
        }

        if mailbox != nil {
            msg, _ := mailbox.RecieveMessage(time.Duration(100*time.Millisecond))
            if msg != nil {
                msgString, err := json.Marshal(msg)
                if err != nil {
                    canolog.Error("Unexpected error: ", err)
                }
                
                canolog.Websocket("Websocket sending: ", msgString)
                websocket.Message.Send(ws, msgString)
            }
        }
    }
}
