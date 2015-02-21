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
    "canopy/datalayer"
    "fmt"
)

/*func (pigeon *CanopyPigeon) Listen(
    name string, 
    func acceptFunc(payload map[string]interface{}) (response map[string]interface{}, error)) {
}

func (pigeon *CanopyPigeon) Launch(name string, payload map[string]inteface{}) {
    // Launch picks a single listener to recieve the job

    // For now, pick a random listener among the candidates to send the paylaod
    // to.
    candidates := pigeon.FilterListeners(string)
    if len(candidates) == 0 {
        return fmt.Errorf("Pigeon: No listeners found for %s", name)
    }
    listenerIdx = rand.Intn(len(candidates))
    
    candidate.Receive()
}
*/

type PigeonSystem struct {
    dl datalayer.PigeonSystem
}

func (pigeon *PigeonSystem) StartWorker(hostname string) (Worker, error) {
    err := pigeon.dl.RegisterWorker(hostname)
    if err != nil {
        return nil, err
    }

    worker := &PigeonWorker{
        hostname: hostname,
    }

    return worker, err
}

func (pigeon *PigeonSystem) Worker(hostname string) (Worker, error) {
    worker := &PigeonWorker{
        hostname: hostname,
    }

    return worker, nil
}

func (pigeon *PigeonSystem) Workers() ([]Worker, error) {
    return nil, fmt.Errorf("Not Implemented")
}
