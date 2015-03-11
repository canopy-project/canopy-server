// Copyright 2015 Gregory Prisament
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

package jobqueue

import (
    "canopy/canolog"
    "encoding/gob"
    "fmt"
    "net"
    "net/rpc"
    "net/http"
    "math/rand"
    "runtime"
)

type PigeonServer struct {
    sys *PigeonSystem
    hostname string
    
    // mapping from msgKey to list of inboxes
    inboxesByMsgKey map[string]([]*PigeonInbox)
}

type pigeonHandler struct {
    fn HandlerFunc
    userCtx map[string]interface{}
}

// RPC entrypoint
func (server *PigeonServer) rpcHandleRequest(req *PigeonRequest, resp *PigeonResponse) (outErr error) {

    // Log crashes in the RPC code
    defer func() {
        r := recover()
        if r != nil {
            var buf [4096]byte
            runtime.Stack(buf[:], false)
            canolog.Error("RPC PANIC ", r, string(buf[:]))
            canolog.Info("Recovered")
            outErr = fmt.Errorf("Crash in %s", req.ReqJobKey)
        }
    }()

    canolog.Info("RPC Handling", req.ReqJobKey)

    // Lookup the handler for that job type
    inboxes, ok := server.inboxesByMsgKey[req.ReqJobKey]
    if !ok {
        // NOT FOUND (NO INBOX LIST)
        return fmt.Errorf("Pigeon Server: No inbox for msg key %s on server %s", req.ReqJobKey, server.hostname)
    }
    if len(inboxes) < 0 {
        // NOT FOUND (NO INBOXES IN LIST)
        return fmt.Errorf("Pigeon Server: No inboxes for msg key %s on server %s", req.ReqJobKey, server.hostname)
    }
    // TODO: handle broadcast & idempotent request
    // For now, send to random inbox

    inbox := inboxes[rand.Intn(len(inboxes))]

    if inbox.handler == nil {
        return fmt.Errorf("Pigeon Server: Expected handler for inbox %s on inbox %s", req.ReqJobKey, inbox)
    }

    // Call the handler
    canolog.Info("Calling Registered handler")
    canolog.Info(req)
    canolog.Info(resp)
    canolog.Info("inbox: ", inbox)
    inbox.handler.Handle(req.ReqJobKey, inbox.userCtx, req, resp)
    canolog.Info("All done")

    return nil
}

func (server *PigeonServer) RPCHandleRequest(req *PigeonRequest, resp *PigeonResponse) error {
    // defer does not seem to work correctly inside main RPC routine.  So this
    // is our workaround.
    err := server.rpcHandleRequest(req, resp) 
    canolog.Info("Leaving RPCHandleRequest")
    return err
}

func (server *PigeonServer) serveRPC() error {
    // TODO: Use direct TCP instead of HTML
    gob.Register(map[string]interface{}{})
    gob.Register(map[string]string{})
    PIGEON_RPC_PORT := ":1888"
    err := rpc.Register(server)
    if err != nil {
        return err
    }
    rpc.HandleHTTP()
    l, err := net.Listen("tcp", PIGEON_RPC_PORT)
    if err != nil {
        return err
    }
    go http.Serve(l, nil)
    return nil
}

func (server *PigeonServer) CreateInbox(msgKey string) (Inbox, error) {
    // Create new inbox object
    inbox := &PigeonInbox{
        server: server,
        msgKey: msgKey,
    }

    // Register this inbox (ie "listener") in the DB
    err := server.sys.dl.RegisterListener(server.hostname, msgKey)
    if err != nil {
        return nil, err
    }

    // Associate the inbox with the msgKey (locally)
    _, ok := server.inboxesByMsgKey[msgKey]
    if ok {
        // Append new inbox to the list
        server.inboxesByMsgKey[msgKey] = append(server.inboxesByMsgKey[msgKey], inbox)
    } else {
        // This is the first inbox on this server for msgKey.
        // Create list of inboxes for msgKey.
        server.inboxesByMsgKey[msgKey] = []*PigeonInbox{inbox}
    }

    return inbox, nil

}

func (server *PigeonServer) Start() error {
    err := server.sys.dl.RegisterWorker(server.hostname)
    if err != nil {
        return err
    }

    err = server.serveRPC()
    if err != nil {
        // TODO: unregister?
        return err
    }

    return nil
}

func (server *PigeonServer) Status() (StatusEnum, error) {
    return DOES_NOT_EXIST, fmt.Errorf("Not implemented")
}

func (server *PigeonServer) Stop() error {
    return fmt.Errorf("Not implemented")
}

func (server *PigeonServer) StopHandling(jobKey string) error {
    return fmt.Errorf("Not implemented")
}
