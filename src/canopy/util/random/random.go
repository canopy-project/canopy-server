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
package random

import (
    "encoding/base64"
    cryptorand "crypto/rand"
    mathrand "math/rand"
)

func Base64String(numChars int) (string, error) {
    randBytes := make([]byte, numChars)
    _, err := cryptorand.Read(randBytes)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(randBytes), nil
}

// Given a slice, returns a random selection of elements from the slice,
// without replacement.
//
// The number of items returned is:
//      MIN(count, len(in))
func Selection(in []interface{}, count uint32) []interface{} {
    out := []interface{}{}

    perm := mathrand.Perm(len(in))
    
    for i := uint32(0); i < count && i < uint32(len(in)); i++ {
        out = append(out, in[perm[i]])
    }
    return out
}

func SelectionStrings(in []string, count uint32) []string {
    out := []string{}

    perm := mathrand.Perm(len(in))
    
    for i := uint32(0); i < count && i < uint32(len(in)); i++ {
        out = append(out, in[perm[i]])
    }
    return out
}
