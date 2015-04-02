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

// Abstract interface that all canopy-op commands must implement

import (
    "canopy/config"
)

type Command interface {
    // Print help for this command
    Help()

    // Get short description of command
    HelpOneLiner() string

    // Returns true if this command object handles <cmdString>.
    Match(cmdString string) bool

    // Carry out the command
    Perform(info CommandInfo)
}

type CommandInfo struct {
    CmdList []Command
    Cfg config.Config
    Args []string
}

// Given a list of command objects, finds the first one that handles
// <cmdString>, if any, or returns nil if no match is found.
func FindCommand(cmds []Command, cmdString string) Command {
    for _, cmd := range cmds {
        if cmd.Match(cmdString) {
            return cmd
        }
    }
    return nil
}
