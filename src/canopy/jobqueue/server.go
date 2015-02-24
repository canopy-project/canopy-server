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
    "encoding/gob"
    "fmt"
    "net"
    "net/rpc"
    "net/http"
)

type PigeonServer struct {
    sys *PigeonSystem
    hostname string
    
    // mapping from jobKey to HandlerFunc
    handlers map[string]*pigeonHandler
}

type PigeonRequest struct {
    ReqJobKey string
    ReqBody map[string]interface{}
}

type PigeonResponse struct {
    RespErr error
    RespBody map[string]interface{}
}

type pigeonHandler struct {
    fn HandlerFunc
    userCtx map[string]interface{}
}

// RPC entrypoint
func (server *PigeonServer) RPCHandleRequest(req *PigeonRequest, resp *PigeonResponse) error {

    // Lookup the handler for that job type
    handler, ok := server.handlers[req.ReqJobKey]
    if !ok {
        // NOT FOUND
        return fmt.Errorf("Pigeon Server: No handler for job key %s on server %s", req.ReqJobKey, server.hostname)
    }

    // Call the handler
    handler.fn(handler.userCtx, req, resp)

    return nil
}

func (server *PigeonServer) serveRPC() error {
    // TODO: Use direct TCP instead of HTML
    gob.Register(map[string]interface{}{})
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

func (server *PigeonServer) Handle(jobKey string, fn HandlerFunc, userCtx map[string]interface{}) error {
    // Register this handler in the DB
    err := server.sys.dl.RegisterListener(server.hostname, jobKey)
    if err != nil {
        return err
    }

    // Associate the handler function with the jobKey
    server.handlers[jobKey] = &pigeonHandler{
        fn: fn,
        userCtx: userCtx,
    }
    return nil
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

func (server *PigeonServer) Status() error {
    return fmt.Errorf("Not implemented")
}

func (server *PigeonServer) Stop() error {
    return fmt.Errorf("Not implemented")
}

func (server *PigeonServer) StopHandling(jobKey string) error {
    return fmt.Errorf("Not implemented")
}
