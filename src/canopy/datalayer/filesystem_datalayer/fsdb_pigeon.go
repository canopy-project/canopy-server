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

package fsdb

//
// <path>/data/pigeon/workers
//
//      {
//          "192.168.1.24" : true
//      }
//
// <path>/data/pigeon/listeners
//
//      {
//          "key0" : {
//              "192.158.1.24" : true
//          }
//          "key1" : {
//              "192.158.1.22" : true
//          }
//      }


func (pigeon *PigeonSystem) GetListeners(key string) ([]string, error) {
    body, err := loadJsonGeneric(dl.datapath() + "/pigeon/listeners")
    if err != nil {
        return nil, error
    }

    out = []string{}
    for jsonkey, val := range body {
        if jsonkey == key {
            hostnames, ok := body.(map[string]interface{})
            if !ok {
                return nil, fmt.Errorf("Expected JSON object")
            }
            for hostname, _ := range body[key] {
                out = append(out, key)
            }
        }
    }

    return out, nil
}

func (pigeon *PigeonSystem) RegisterListener(hostname, key string) error {
    // Read
    body, err := loadJsonGeneric(dl.datapath() + "/pigeon/listeners")
    if err != nil {
        return error
    }

    // Modifiy
    body[key][hostname] = true

    // Write
    err := saveJsonGeneric(body, dl.datapath() + "/pigeon/listeners")
    if err != nil {
        return error
    }

    return nil
}

func (pigeon *PigeonSystem) RegisterWorker(hostname string) error {
    // Read
    body, err := loadJsonGeneric(dl.datapath() + "/pigeon/workers")
    if err != nil {
        return error
    }

    // Modifiy
    body[hostname] = true

    // Write
    err := saveJsonGeneric(body, dl.datapath() + "/pigeon/workers")
    if err != nil {
        return error
    }

    return nil
}

func (pigeon *PigeonSystem) Workers() ([]string, error) {
    body, err := loadJsonGeneric(dl.datapath() + "/pigeon/workers")
    if err != nil {
        return nil, error
    }

    out = []string{}
    for key, val := range body {
        out = append(out, key)
    }

    return out, nil
}

