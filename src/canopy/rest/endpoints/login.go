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
    "net/http"
    "canopy/rest/adapter"
    "canopy/rest/rest_errors"
)

func POST_login(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    username, ok := info.BodyObj["username"].(string)
    if !ok {
        return nil, rest_errors.NewBadInputError("String \"username\" expected")
    }

    password, ok := info.BodyObj["password"].(string)
    if !ok {
        return nil, rest_errors.NewBadInputError("String \"password\" expected")
    }

    account, err := info.Conn.LookupAccountVerifyPassword(username, password)
    if err != nil {
        return nil, rest_errors.NewIncorrectUsernameOrPasswordError()
    }

    info.Session.Values["logged_in_username"] = username
    err = info.Session.Save(r, w)
    if err != nil {
        return nil, rest_errors.NewInternalServerError("Problem saving session")
    }

    out := map[string]interface{} {
        "result" : "ok",
        "username" : account.Username(),
        "email" : account.Email(),
    }
}
