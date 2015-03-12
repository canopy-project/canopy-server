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
    "canopy/mail/messages"
)

// Constructs the response body for the /api/create_account REST endpoint
func ApiCreateUserHandler(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    username, ok := info.BodyObj["username"].(string)
    if !ok {
        return nil, BadInputError("String \"username\" expected").Log()
    }

    email, ok := info.BodyObj["email"].(string)
    if !ok {
        return nil, BadInputError("String \"email\" expected").Log()
    }

    password, ok := info.BodyObj["password"].(string)
    if !ok {
        return nil, BadInputError("String \"password\" expected").Log()
    }

    skipEmail := false
    skipEmailItf, ok := info.BodyObj["skip-email"]
    if ok {
        skipEmail, ok = skipEmailItf.(bool)
        if !ok {
            return nil, BadInputError("\"skip-email\" must be boolean").Log()
        }
    }

    account, err := info.Conn.LookupAccount(username)
    if err == nil {
        // TODO: other errors could have occurred.  Do not necessarily take
        // (err != nil) as a sign that username is available!
        return nil, UsernameNotAvailableError().Log()
    }

    account, err = info.Conn.LookupAccount(email)
    if err == nil {
        // TODO: other errors could have occurred.  Do not necessarily take
        // (err != nil) as a sign that username is available!
        return nil, EmailTakenError().Log()
    }

    account, err = info.Conn.CreateAccount(username, email, password)
    if err != nil {
        return nil, InternalServerError("Problem Creating Account" + err.Error()).Log()
    }

    sideEffect.SetCookie("logged_in_username", username)

    protocol := "http://"
    if info.Config.OptEnableHTTPS() {
        protocol = "https://"
    }

    activationLink := protocol + info.Config.OptHostname() + 
            "/mgr/activate.html?username=" + account.Username() + 
            "&code=" + account.ActivationCode()

    // Send email
    if !skipEmail {
        msg := sideEffect.SendEmail()
        msg.AddTo(account.Email(), account.Username())
        msg.SetFrom("no-reply@canopy.link", "Canopy Cloud Service")
        msg.SetReplyTo("no-reply@canopy.link")
        messages.MailMessageCreatedAccount(msg,
            account.Username(), 
            activationLink,
            protocol + info.Config.OptHostname(),
            info.Config.OptHostname(),
        )
    }

    out := map[string]interface{} {
        "activated" : false,
        "result" : "ok",
        "username" : account.Username(),
        "email" : account.Email(),
    }
    return out, nil
}
