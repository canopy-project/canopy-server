// Copright 2015 Canopy Services, Inc.
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

package canopy_ops

// canopy-ops workers
// List pigeon workers

import (
    "canopy/datalayer/cassandra_datalayer"
    "fmt"
)

type WorkersCommand struct{}

func (WorkersCommand)HelpOneLiner() string {
    return "    workers     List all registered canopy workers"
}

func (WorkersCommand)Help() {
    fmt.Println("COMMAND:")
    fmt.Println("   canopy-ops workers")
    fmt.Println("")
    fmt.Println("DESCRIPTION:")
    fmt.Println("   Lists hostnames of all registered canopy workers")
    fmt.Println("")
}

func (WorkersCommand)Match(cmdString string) bool {
    return (cmdString == "workers")
}

func (WorkersCommand)Perform(info CommandInfo) {
    dl := cassandra_datalayer.NewDatalayer(info.Cfg)
    conn, err := dl.Connect()
    if err != nil {
        fmt.Println(err)
    }
    workers, err := conn.PigeonSystem().Workers()
    if err != nil {
        fmt.Println(err)
    }
    for _, worker := range workers {
        fmt.Println(worker)
    }
}
