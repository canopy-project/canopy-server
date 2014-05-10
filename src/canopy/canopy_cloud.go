package main

import (
    "fmt"
    "time"
    "encoding/json"
    "code.google.com/p/go.net/websocket"
    "io"
    "net"
    "canopy/datalayer"
)

// Process JSON message from the client
func processPayload(dl *datalayer.CassandraDatalayer, payload string) {
    var f interface{}
    var deviceId string
    var cpu float64

    err := json.Unmarshal([]byte(payload), &f)
    if err != nil{
        fmt.Println("Error JSON decoding payload: ", payload)
        return;
    }

    m := f.(map[string]interface{})
    for k, v := range m {
        switch vv := v.(type) {
            case float64:
                cpu = vv
            case string:
                deviceId = vv
            default:
                fmt.Println(k, "is of a type I don't know how to handle");
        }
    }

    dl.StorePropertyValue(deviceId, "cpu", cpu)
}

// Main websocket server routine.
// This event loop runs until the websocket connection is broken.
func CanopyWebsocketServer(ws *websocket.Conn) {

    // connect to cassandra
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect()

    for {
        var in string

        // check for message from client
        ws.SetReadDeadline(time.Now().Add(100*time.Millisecond))
        err := websocket.Message.Receive(ws, &in)
        if err == nil {
            // success, payload received
            processPayload(dl, in)
        } else if err == io.EOF {
            // connection closed
            return;
        } else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
            // timeout reached, no data for me this time
        } else {
            fmt.Println("Unexpected error:", err);
        }

        //websocket.Message.Send(ws, message)
    }
}
