// Copright 2014-2015 Canopy Services, Inc.
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

// canopy-ops help [<topic_or_cmd>]

import (
    "fmt"
)

type HelpCommand struct{}

func (HelpCommand)HelpOneLiner() string {
    return "    help        Display help for a topic or command."
}

func (HelpCommand)Help() {
    fmt.Println("Sounds like you really need some help!")
}

func (HelpCommand)Match(cmdString string) bool {
    return (cmdString == "help")
}

func (HelpCommand)Perform(info CommandInfo) {
    if len(info.Args) == 2 {
        cmd := FindCommand(info.CmdList, info.Args[1])
        if cmd != nil {
            cmd.Help()
        } else {
            fmt.Println("No manual entry for", info.Args[1])
        }
    } else {
        fmt.Println("Usage: canopy-ops <cmd> [<args>]")
        fmt.Println("")
        fmt.Println("Commands:")
        for _, cmd := range info.CmdList {
            fmt.Println(cmd.HelpOneLiner())
        }
    }
}
