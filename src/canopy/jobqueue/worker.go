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
func (worker *PigeonWorker) HandleRequest(request *PigeonRequest, response *PigeonResponse) error {
    canolog.Info("Pigeon Worker RPC: Request recieved: ", request)
    // Lookup the listener for that job type
    listener, ok := worker.listeners[request.ReqKey]
    if !ok {
        // NOT FOUND
        return fmt.Errorf("Pigeon Worker: No handler for job key %s on worker %s", request.ReqKey, worker.hostname)
    }
    canolog.Info("Pigeon Worker RPC: Listener found")

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
    canolog.Info("REgistering Worker")
    rpc.Register(worker)
    canolog.Info("Handling HTTP")
    rpc.HandleHTTP()
    canolog.Info("Listening on port 1888")
    l, err := net.Listen("tcp", PIGEON_RPC_PORT)
    if err != nil {
        canolog.Error(err)
        return err
    }
    canolog.Info("Start Serving")
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

func (worker *PigeonWorker) Start() error {
    canolog.Info("Starting Worker")
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
