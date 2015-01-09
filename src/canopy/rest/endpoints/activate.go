// Copyright 2014 SimpleThings, Inc.
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
package endpoints

import (
    "canopy/rest/adapter"
    "canopy/rest/rest_errors"
    "net/http"
)

// Request:
// {
//     "username" : 
//     "code" : 
// }
func POST_activate(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    if info.Account == nil {
        return nil, rest_errors.NewNotLoggedInError()
    }

    username, ok := info.BodyObj["username"].(string)
    if !ok {
        return nil, rest_errors.NewBadInputError("String \"username\" expected")
    }

    code, ok := info.BodyObj["code"].(string)
    if !ok {
        return nil, rest_errors.NewBadInputError("String \"activation_code\" expected")
    }

    err := info.Account.Activate(username, code)
    if err != nil {
        // TODO: Report InternalServerError different from InvalidCode
        return nil, rest_errors.NewBadInputError("Unable to activate account")
    }

    return map[string]interface{} {
        "result" : "ok",
    }, nil
}

