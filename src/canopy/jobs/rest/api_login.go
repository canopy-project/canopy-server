// Copyright 2014-2015 Canopy Services, Inc.
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

func POST_api__login(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    username, ok := info.BodyObj["username"].(string)
    if !ok {
        return nil, BadInputError("String \"username\" expected")
    }

    password, ok := info.BodyObj["password"].(string)
    if !ok {
        return nil, BadInputError("String \"password\" expected")
    }

    account, err := info.Conn.LookupAccountVerifyPassword(username, password)
    if err != nil {
        return nil, IncorrectUsernameOrPasswordError()
    }

    sideEffect.SetCookie("logged_in_username", username)

    out := map[string]interface{} {
        "result" : "ok",
        "username" : account.Username(),
        "email" : account.Email(),
    }
    return out, nil
}
