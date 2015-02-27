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
)

// Constructs the response body for the /api/me REST endpoint
func GET__api__me(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, NotLoggedInError().Log()
    }
    return map[string]interface{}{
        "activated" : info.Account.IsActivated(),
        "email" : info.Account.Email(),
        "result" : "ok",
        "username" : info.Account.Username(),
    }, nil
}


func POST__api__me(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, NotLoggedInError()
    }
    for fieldName, value := range info.BodyObj {
        switch fieldName {
        case "email":
            return nil, InternalServerError("Changing email not implemented")
        case "new_password":
            newPassword, ok := value.(string)
            if !ok {
                return nil, BadInputError("Expected string \"new_password\"")
            }
            oldPasswordObj, ok := info.BodyObj["old_password"]
            if !ok {
                return nil, BadInputError("Must provide \"old_password\" to change password")
            }
            oldPassword, ok := oldPasswordObj.(string)
            if !ok {
                return nil, BadInputError("Expected string \"old_password\"")
            }
            ok = info.Account.VerifyPassword(oldPassword);
            if (!ok) {
                return nil, BadInputError("Incorrect old password")
            }
            err := info.Account.SetPassword(newPassword)
            if err != nil {
                // TODO: finer-grained error reporting
                return nil, InternalServerError("Problem changing password")
            }
        }
    }
    return map[string]interface{}{
        "result" : "ok",
        "username" : info.Account.Username(),
        "email" : info.Account.Email(),
    }, nil
}
