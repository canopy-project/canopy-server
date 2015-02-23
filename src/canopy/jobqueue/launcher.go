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
    "fmt"
    "net/rpc"
    "math/rand"
    "canopy/util/random"
)

type PigeonLauncher struct {
    sys *PigeonSystem
    timeoutms uint32
}

func (launcher *PigeonLauncher) send(hostname string, payload map[string]interface{}) error {
    response := PigeonResponse{}

    // Dial the server
    client, err := rpc.DialHTTP("tcp", hostname)
    if err != nil {
        return fmt.Errorf("Pigeon: (dialing) %s", err.Error())
    }
    defer client.Close()

    // Make the call
    client.Call("PigeonRPC.Request", payload, &response)
    if err != nil {
        return fmt.Errorf("Pigeon: (calling) %s", err.Error())
    }
    
    return nil
}

func (launcher *PigeonLauncher) Broadcast(key string, payload map[string]interface{}) error {
    // Get list of all workers interested in these keys
    workerHosts, err := launcher.sys.dl.GetListeners(key)
    if err != nil {
        return err
    }

    // Send message to each worker
    for _, workerHost := range workerHosts {
        launcher.send(workerHost, payload)
    }

    return nil
}

func (launcher *PigeonLauncher) Launch(key string, payload map[string]interface{}) (<-chan PigeonResponse, error) {
    // Get list of all workers interested in these keys
    workerHosts, err := launcher.sys.dl.GetListeners(key)
    if err != nil {
        return nil, err
    }

    if len(workerHosts) == 0 {
        return nil, fmt.Errorf("Pigeon: No listeners found for %s", key)
    }

    // For now, pick one at random
    workerHost := workerHosts[rand.Intn(len(workerHosts))]

    launcher.send(workerHost, payload)
    return nil, nil
}

func (launcher *PigeonLauncher) LaunchIdempotent(key string, numParallel uint32, payload map[string]interface{}) (<-chan PigeonResponse, error) {
    // Get list of all workers interested in these keys
    workerHosts, err := launcher.sys.dl.GetListeners(key)
    if err != nil {
        return nil, err
    }

    if len(workerHosts) == 0 {
        return nil, fmt.Errorf("Pigeon: No listeners found for %s", key)
    }

    // For now, pick a random subset of numParallel workers
    workerHostsSubset := random.SelectionStrings(workerHosts, numParallel)

    // Send payload to each of the workers
    for _, worker := range workerHostsSubset {
        // TODO: take response of first responder
        launcher.send(worker, payload)
    }

    return nil, fmt.Errorf("Not fully implemented")
}

func (launcher *PigeonLauncher) SetTimeoutms(timeout uint32) {
    launcher.timeoutms = timeout
}

