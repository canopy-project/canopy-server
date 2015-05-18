/*
 * Copyright 2014-2015 Canopy Services, Inc.
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

package ws

import (
    "time"
    "encoding/json"
    "code.google.com/p/go.net/websocket"
    "io"
    "net"
    "canopy/canolog"
    "canopy/config"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/pigeon"
    "canopy/service"
)


/*func IsDeviceConnected(pigeonSys *pigeon.PigeonSystem, deviceIdString string) bool {
    return (pigeonSys.Mailbox(deviceIdString) != nil)
}*/

func NewCanopyWebsocketServer(cfg config.Config, outbox jobqueue.Outbox, pigeonServer jobqueue.Server) func(ws *websocket.Conn) {
    // Main websocket server routine.
    // This event loop runs until the websocket connection is broken.
    return func(ws *websocket.Conn) {
        canolog.Websocket("Websocket connection established")

        var cnt int32
        var device datalayer.Device
        var inbox jobqueue.Inbox
        var inboxReciever jobqueue.RecieveHandler
        lastPingTime := time.Now()
        
        cnt = 0

        // connect to cassandra
        dl := cassandra_datalayer.NewDatalayer(cfg)
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
                resp := service.ProcessDeviceComm(cfg, conn, device, "", "", in)
                if resp.Device == nil{
                    canolog.Error("Error processing device communications: ", resp.Err)
                } else {
                    device = resp.Device
                    if inbox == nil {
                        inbox, err = pigeonServer.CreateInbox("canopy_ws:" + device.ID())
                        if err != nil {
                            canolog.Error("Error initializing inbox:", err)
                            return
                        }
                        inboxReciever = jobqueue.NewRecieveHandler()
                        inbox.SetHandler(inboxReciever)

                        err = device.UpdateWSConnected(true)
                        if err != nil {
                            canolog.Error("Unexpected error: ", err)
                        }
                    }
                }
            } else if err == io.EOF {
                canolog.Websocket("Websocket connection closed")
                // connection closed
                if inbox != nil {
                    if device != nil {
                        err = device.UpdateWSConnected(false)
                        if err != nil {
                            canolog.Error("Unexpected error: ", err)
                        }
                    }
                    inbox.Close()
                }
                return;
            } else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
                // timeout reached, no data for me this time
            } else {
                canolog.Error("Unexpected error: ", err)
            }

            // Periodically send blank message
            if time.Now().After(lastPingTime.Add(30*time.Second)) {
                err := websocket.Message.Send(ws, "{}")
                if err != nil {
                    canolog.Websocket("Websocket connection closed during ping")
                    // connection closed
                    if inbox != nil {
                        if device != nil {
                            err = device.UpdateWSConnected(false)
                            if err != nil {
                                canolog.Error("Unexpected error: ", err)
                            }
                        }
                        inbox.Close()
                    }
                    return;
                }
                canolog.Info("Pinging WS")
                lastPingTime = time.Now()
            }

            if inbox != nil {
                msg, _ := inboxReciever.Recieve(time.Duration(100*time.Millisecond))
                if msg != nil {
                    msgString, err := json.Marshal(msg)

                    if err != nil {
                        canolog.Error("Unexpected error: ", err)
                    }
                    
                    canolog.Info("Websocket sending", msgString)
                    canolog.Websocket("Websocket sending: ", msgString)
                    websocket.Message.Send(ws, msgString)
                }
            }
        }
    }
}
