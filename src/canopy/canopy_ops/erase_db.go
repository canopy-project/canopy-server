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

// canopy-ops erase-db
// Wipes database

import (
    "canopy/datalayer/cassandra_datalayer"
    "fmt"
)

type EraseDBCommand struct{}

func (EraseDBCommand)HelpOneLiner() string {
    return "    erase-db    Wipe entire database"
}

func (EraseDBCommand)Help() {
    fmt.Println("COMMAND:")
    fmt.Println("   canopy-ops erase-db")
    fmt.Println("")
    fmt.Println("DESCRIPTION:")
    fmt.Println("   Wipes entire database.  Use with caution!")
    fmt.Println("")
}

func (EraseDBCommand)Match(cmdString string) bool {
    return (cmdString == "erase-db")
}

func (EraseDBCommand)Perform(info CommandInfo) {
    dl := cassandra_datalayer.NewDatalayer(info.Cfg)
    dl.EraseDb()
}
