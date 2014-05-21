package main

import (
    "fmt"
    "github.com/gocql/gocql"
    "time"
    "encoding/json"
    "code.google.com/p/go.net/websocket"
    "io"
    "net"
    "canopy/datalayer"
    "canopy/pigeon"
)

// Process JSON message from the client
func processPayload(dl *datalayer.CassandraDatalayer, payload string) string{
    var f interface{}
    var deviceIdString string
    var cpu float64

    err := json.Unmarshal([]byte(payload), &f)
    if err != nil{
        fmt.Println("Error JSON decoding payload: ", payload)
        return "";
    }

    m := f.(map[string]interface{})
    for k, v := range m {
        switch vv := v.(type) {
            case float64:
                cpu = vv
            case string:
                deviceIdString = vv
            default:
                fmt.Println(k, "is of a type I don't know how to handle");
        }
    }

    deviceId, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        fmt.Println("Invalid UUID", deviceIdString, err);
        return ""
    }

    device, err := dl.LookupDevice(deviceId)
    if err != nil {
        fmt.Println("Could not lookup device: ", deviceIdString, err)
        return ""
    }
    err = device.InsertSensorSample("cpu", time.Now(), cpu);
    if err != nil {
        fmt.Println("Error saving sample", err)
        return ""
    }

    return deviceIdString;
}

// Main websocket server routine.
// This event loop runs until the websocket connection is broken.
func CanopyWebsocketServer(ws *websocket.Conn) {

    var mailbox *pigeon.PigeonMailbox

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
            deviceId := processPayload(dl, in)
            if deviceId != "" && mailbox == nil {
                mailbox = gPigeon.CreateMailbox(deviceId)
            }
        } else if err == io.EOF {
            // connection closed
            return;
        } else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
            // timeout reached, no data for me this time
        } else {
            fmt.Println("Unexpected error:", err);
        }

        if mailbox != nil {
            msg, _ := mailbox.RecieveMessage(time.Duration(100*time.Millisecond))
            if msg != nil {
                fmt.Println("Message recieved: ", msg)
            }
        }

        //websocket.Message.Send(ws, message)
    }
}
