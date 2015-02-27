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

package rest

import (
    "canopy/canolog"
)

// Backend implementation /api/activate endpoint
// Activates a user account (i.e., email address confirmation).
// 
func ApiActivateHandler(info *RestRequestInfo) (map[string]interface{}, RestError) {
    canolog.Info("api/activate REST job started")
    if info.Account == nil {
        return nil, NotLoggedInError().Log()
    }

    username, ok := info.BodyObj["username"].(string)
    if !ok {
        return nil, BadInputError(`String "username" expected`).Log()
    }

    code, ok := info.BodyObj["code"].(string)
    if !ok {
        return nil, BadInputError(`String "code" expected`).Log()
    }

    err := info.Account.Activate(username, code)
    if err != nil {
        // TODO: Report InternalServerError different from InvalidCode
        //return nil, rest_errors.NewBadInputError("Unable to activate account")
        return nil, BadInputError("Unable to activate account").Log()
    }

    return map[string]interface{}{
        "result" : "ok",
    }, nil
}
