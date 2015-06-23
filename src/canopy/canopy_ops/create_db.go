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

// canopy-ops create-db
// Initialize database

import (
    "canopy/datalayer/cassandra_datalayer"
    "fmt"
)

type CreateDBCommand struct{}

func (CreateDBCommand)HelpOneLiner() string {
    return "    create-db   Initialize database"
}

func (CreateDBCommand)Help() {
    fmt.Println("COMMAND:")
    fmt.Println("   canopy-ops create-db")
    fmt.Println("")
    fmt.Println("DESCRIPTION:")
    fmt.Println("   Initialize database")
    fmt.Println("")
}

func (CreateDBCommand)Match(cmdString string) bool {
    return (cmdString == "create-db")
}

func (CreateDBCommand)Perform(info CommandInfo) {
    fmt.Println("Running create-db:")
    dl := cassandra_datalayer.NewDatalayer(info.Cfg)
    err := dl.PrepDb(3) // 3 = replication factor
    if err != nil {
        fmt.Println(err)
    }
}
