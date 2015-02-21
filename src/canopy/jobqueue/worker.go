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
)

type PigeonWorker struct {
    sys *PigeonSystem
    hostname string
}

func (worker *PigeonWorker) Listen(key string, request <-chan Request, response <-chan Response) error {
    return fmt.Errorf("Not implemented")
}

func (worker *PigeonWorker) Start() error {
    return fmt.Errorf("Not implemented")
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
