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
        sddlClass, err := sddl.ParseClass("anonymous", sddlJson)
        if !ok {
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
    if cnt % 10 == 0 {
        for k, v := range payloadObj {
            /* hack */
            if k == "device_id" || k == "sddl" {
                continue
            }
            switch vv := v.(type) {
                case float64:
                    err = device.InsertSensorSample(k, time.Now(), vv);
                    if err != nil {
                        fmt.Println("Error saving sample", err)
                        return ""
                    }
                default:
                    fmt.Println(k, "is of a type I don't know how to handle");
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
