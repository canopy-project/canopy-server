/*
 * Copyright 2014 Gregory Prisament
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
package endpoints

import (
    "canopy/rest/adapter"
    "canopy/rest/rest_errors"
    "net/http"
)

func GET_me(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    if info.Account == nil {
        return nil, rest_errors.NewNotLoggedInError()
    }
    return map[string]interface{}{
        "result" : "ok",
        "username" : info.Account.Username(),
        "email" : info.Account.Email(),
    }, nil
}

func POST_me(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    if info.Account == nil {
        return nil, rest_errors.NewNotLoggedInError()
    }

    for fieldName, value := range info.BodyObj {
        switch fieldName {
        case "email":
            return nil, rest_errors.NewInternalServerError("Changing email not implemented")
        case "new_password":
            newPassword, ok := value.(string)
            if !ok {
                return nil, rest_errors.NewBadInputError("Expected string \"new_password\"")
            }
            oldPasswordObj, ok := info.BodyObj["old_password"]
            if !ok {
                return nil, rest_errors.NewBadInputError("Must provide \"old_password\" to change password")
            }
            oldPassword, ok := oldPasswordObj.(string)
            if !ok {
                return nil, rest_errors.NewBadInputError("Expected string \"old_password\"")
            }
            ok = info.Account.VerifyPassword(oldPassword);
            if (!ok) {
                return nil, rest_errors.NewBadInputError("Incorrect old password")
            }

            err := info.Account.SetPassword(newPassword)
            if err != nil {
                // TODO: finer-grained error reporting
                return nil, rest_errors.NewInternalServerError("Problem changing password")
            }
        }
    }

    return map[string]interface{}{
        "result" : "ok",
        "username" : info.Account.Username(),
        "email" : info.Account.Email(),
    }, nil
}

