// Copyright 2015 Canopy Services, Inc.
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
    "net/rpc"
    "math/rand"
    //"canopy/util/random"
)

type PigeonOutbox struct {
    sys *PigeonSystem
    timeoutms int32
}

func (outbox *PigeonOutbox) send(hostname string, request *PigeonRequest, respChan chan<- Response) error {
    resp := &PigeonResponse{}

    // Dial the server
    // TODO: Inefficient to dial each time?
    canolog.Info("RPC Dialing")
    rpcClient, err := rpc.DialHTTP("tcp", hostname + ":1888")
    if err != nil {
        return fmt.Errorf("Pigeon: (dialing) %s", err.Error())
    }
    defer rpcClient.Close()

    // Make the call
    canolog.Info("RPC Calling")
    err = rpcClient.Call("PigeonServer.RPCHandleRequest", request, resp)
    if err != nil {
        canolog.Error("Pigeon: (calling) ", err.Error())
        // Send error response to channel
        respChan <- resp
        return fmt.Errorf("Pigeon: (calling) %s", err.Error())
    }

    // Send response to channel
    respChan <- resp
    
    return nil
}

func (outbox *PigeonOutbox) Broadcast(key string, payload map[string]interface{}) error {
    // Get list of all workers interested in these keys
    //workerHosts, err := launcher.sys.dl.GetListeners(key)
    //if err != nil {
    //    return err
    //}

    // Send message to each worker
    //for _, workerHost := range workerHosts {
        //launcher.send(workerHost, payload)
    //}

    return nil
}

func (outbox *PigeonOutbox) Launch(key string, payload map[string]interface{}) (<-chan Response, error) {
    canolog.Info("Launching ", key)

    req := PigeonRequest {
        ReqJobKey: key,
        ReqBody: payload,
    }

    // Get list of all workers interested in these keys
    serverHosts, err := outbox.sys.dl.GetListeners(key)
    if err != nil {
        return nil, err
    }

    if len(serverHosts) == 0 {
        canolog.Info("No listeners found", key)
        return nil, fmt.Errorf("Pigeon: No listeners found for %s", key)
    }

    // For now, pick one at random
    serverHost := serverHosts[rand.Intn(len(serverHosts))]

    canolog.Info("Making RPC call ", key)
    respChan := make(chan Response)
    go outbox.send(serverHost, &req, respChan)
    canolog.Info("Returned from send", key)

    return respChan, nil
}

func (outbox *PigeonOutbox) LaunchIdempotent(key string, numParallel uint32, payload map[string]interface{}) (<-chan Response, error) {
    // Get list of all workers interested in these keys
    //workerHosts, err := outbox.sys.dl.GetListeners(key)
    //if err != nil {
    //    return nil, err
    //}

    //if len(workerHosts) == 0 {
    //    return nil, fmt.Errorf("Pigeon: No listeners found for %s", key)
    //}

    // For now, pick a random subset of numParallel workers
    //workerHostsSubset := random.SelectionStrings(workerHosts, numParallel)

    // Send payload to each of the workers
    //for _, worker := range workerHostsSubset {
        // TODO: take response of first responder
        //outbox.send(worker, payload)
    //}

    return nil, fmt.Errorf("Not fully implemented")
}

func (outbox *PigeonOutbox) SetTimeoutms(timeout int32) {
    outbox.timeoutms = timeout
}

