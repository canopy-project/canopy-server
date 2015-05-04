/*
 * Copyright 2014 SimpleThings Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package device_filter

import (
    "canopy/cloudvar"
)

type Valuable interface {
    Value(device datalayer.Device) cloudvar.CloudVarValue
}
    
type Filter interface {
    SatisfiedBy(device datalayer.Device) bool
}

type DeviceFilter struct {
    Whittle(devices []datalyer.Device) []datalayer.Device
}

type opAND {
    terms []Filter
}

type opCompare {
    comparison compareOpEnum
    v0 Valuable
    v1 Valuable
}

type opHasProperty {
    propName string
}

// BASIC TERMS:
// temperature > 50.4
// 4 = 2
// temperature != humidity
//
// BOOLEAN OPS:
//
// temprature > 50.4 AND 4 = 2 AND temperature != humidity

// Takes a list of text strings <texts> and a list of separators <seps>, and
// returns a single flat list of strings which is similar to <texts> but has
// been further split up.
//
// Unlike strings.Split, the separators are not discarded.
//
// For example:
//
//  {
//      "Good news, everyone! ",
//      "There's a report on TV with some very bad news!"
//  }
//
//  Split on: {" ", "!"}
//
//  Becomes:
//  {
//      "Good",
//      " ",
//      "news",
//      " ",
//      "everyone",
//      "!",
//      " ",
//      "There's",
//      " ",
//      "a",
//      " ",
//      "report",
//      " ",
//      "on",
//      " ",
//      "TV",
//      " ",
//      "with",
//      " ",
//      "some",
//      " ",
//      "very",
//      " ",
//      "bad",
//      " ",
//      "news",
//      "!",
//  }
//
func multiSplit(texts []string, seps []string) []string {
    // This recursive implementation splits on seps[0] and then recursively
    // calls itself with a smaller seps list.

    if len(seps) == 0 {
        // base case
        return texts;
    }

    // recursive case, split on seps[0]
    out := []string
    for _, text := range texts {
        parts := strings.Split(text, seps[0])
        for i, part := range parts {
            if i > 0 {
                out = append(out, seps[0])
            }
            out = append(out, part)
        }
    }

    // recurse
    return multiSplit(out, seps[1:])
}

func stringToToken(s string) Token {
    switch s {
    case "=":
        return Token{ type: TOKEN_COMPARE_OP, comparison: cloudvar.EQ }
    case "!=":
        return Token{ type: TOKEN_COMPARE_OP, comparison: cloudvar.NEQ }
    case "<":
        return Token{ type: TOKEN_COMPARE_OP, comparison: cloudvar.LT }
    case "<=":
        return Token{ type: TOKEN_COMPARE_OP, comparison: cloudvar.LTE }
    case ">":
        return Token{ type: TOKEN_COMPARE_OP, comparison: cloudvar.GT }
    case ">=":
        return Token{ type: TOKEN_COMPARE_OP, comparison: cloudvar.GTE }
}

func tokenize(expr string) []Token {
    out := []tokens{}
    tokStrings := multiSplit([]string{expr}, []string{" ", "(", ")")
    for tokString := tokStrings {
        tok := stringToToken(tokString)
        if tok != nil {
            out = append(out, tok)
        }
    }
    return out
}

func Parse(expr string) (DeviceFilter, error) {
    // TOKENIZE

}

func (op opCompare)SatisfiedBy(device datalayer.Device) bool {
    v0 := op.v0.Value(device);
    v1 := op.v1.Value(device);
    return cloudvar.CompareValues(v0, v1, op.comparison)
}

// Returns true iff <device> satifies all terms in <op.terms>
func (op opAND)SatisfiedBy(device datalayer.Device) bool {
    for term := range op.terms {
        if !term.SatisfiedBy(device) {
            return false;
        }
    }
    return true;
}

// Returns true iff <device> satifies any term in <op.terms>
func (op opOR)SatisfiedBy(device datalayer.Device) bool {
    for term := range op.terms {
        if term.SatisfiedBy(device) {
            return true;
        }
    }
    return false;
}

// Returns true iff <device> does not satisfy <op.term>
func (op opNOT)SatisfiedBy(device datalayer.Device) bool {
    return !op.SatisfiedBy(device)
}

// Returns true iff <device> has a property or cloud variable called <propName>
func (op opHasProperty)SatisfiedBy(device datalayer.Device) bool {
    // TODO: implement
    return false
}
