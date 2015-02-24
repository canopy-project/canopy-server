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
    "fmt"
    "net"
    "net/rpc"
    "net/http"
)

type PigeonWorker struct {
    sys *PigeonSystem
    hostname string
    listeners map[string]PigeonListener
}

type PigeonRequest struct {
    // job key
    ReqKey string

    // payload
    ReqBody map[string]interface{}
}

type PigeonResponse struct {
    RespErr error
    RespBody map[string]interface{}
}

type PigeonListener struct {
    worker *PigeonWorker
    requestChan chan<- Request
    responseChan <-chan Response
}
func (worker *PigeonWorker) HandleRPCRequest(request *PigeonRequest, response *PigeonResponse) error {
    // Lookup the listener for that job type
    listener, ok := worker.listeners[request.ReqKey]
    if !ok {
        // NOT FOUND
        return fmt.Errorf("Pigeon Worker: No handler for job key %s on worker %s", request.ReqKey, worker.hostname)
    }

    listener.requestChan <- request
    
    // Wait for response
    resp, ok := (<-listener.responseChan).(*PigeonResponse)
    if !ok {
        return fmt.Errorf("Pigeon Worker: Expected PigeonResponse from response handler)")
    }
    response.RespErr = resp.RespErr
    response.RespBody = resp.RespBody

    return nil
}

func (worker *PigeonWorker) serveRPC() error {
    PIGEON_RPC_PORT := ":1888"
    rpc.Register(worker)
    rpc.HandleHTTP()
    l, err := net.Listen("tcp", PIGEON_RPC_PORT)
    if err != nil {
        return err
    }
    go http.Serve(l, nil)
    return nil
}

func (worker *PigeonWorker) Listen(key string, requestChan chan<- Request, responseChan <-chan Response) error {
    err := worker.sys.dl.RegisterListener(worker.hostname, key)
    if err != nil {
        return err
    }

    listener := PigeonListener{
        worker: worker,
        requestChan: requestChan,
        responseChan: responseChan,
    }

    worker.listeners[key] = listener

    return nil
}


func (worker *PigeonWorker) goListenHandlerFunc(key string, requestChan chan Request, responseChan chan Response, func handlerFunc(Request, Response)) {
    for {
        // Wait for request to come in
        req := <-requestChan
        resp := worker.sys.NewResponse()

        // spawn goroutine 
        go func(req Request, resp Response) <-chan bool {
            handlerFunc(req, resp)
        }
        handlerFunc(req, resp)
    }
}

func (worker *PigeonWorker) ListenHandlerFunc(key string, func handlerFunc(Request, Response)) {
    // Create channels
    requestChan := make(chan Request)
    responseChan := make(chan Response)

    worker.Listen(key, requestChan, responseChan)
    go goListenHandlerFunc(key, requestChan, responseChan, handlerFunc)
}

func (worker *PigeonWorker) Start() error {
    err := worker.sys.dl.RegisterWorker(worker.hostname)
    if err != nil {
        return err
    }

    err = worker.serveRPC()
    if err != nil {
        // TODO: unregister?
        return err
    }

    return nil
}

func (worker *PigeonWorker) Status() error {
    return fmt.Errorf("Not implemented")
}

func (worker *PigeonWorker) Stop() error {
    return fmt.Errorf("Not implemented")
}

func (worker *PigeonWorker) StopListening(key string) error {
    return fmt.Errorf("Not implemented")
}
