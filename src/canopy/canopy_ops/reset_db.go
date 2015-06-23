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

// canopy-ops reset-db
// Wipe entire database, then initalize a new database

import (
    "canopy/datalayer/cassandra_datalayer"
    "fmt"
)

type ResetDBCommand struct{}

func (ResetDBCommand)HelpOneLiner() string {
    return "    reset-db    Wipe entire db then initialize a new db"
}

func (ResetDBCommand)Help() {
    fmt.Println("COMMAND:")
    fmt.Println("   canopy-ops reset-db")
    fmt.Println("")
    fmt.Println("DESCRIPTION:")
    fmt.Println("   Equivalent to 'canopy-ops erase-db' followed by 'canopy-ops create-db'.  Use with caution!")
    fmt.Println("")
}

func (ResetDBCommand)Match(cmdString string) bool {
    return (cmdString == "reset-db")
}

func (ResetDBCommand)Perform(info CommandInfo) {
    dl := cassandra_datalayer.NewDatalayer(info.Cfg)
    dl.EraseDb()
    err := dl.PrepDb(3) // 3 = replication factor
    if err != nil {
        fmt.Println(err)
    }
}
