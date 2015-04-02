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

// Constructs the response body for the /api/user/self REST endpoint
func GET__api__user__self(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, NotLoggedInError().Log()
    }
    return map[string]interface{}{
        "validated" : info.Account.IsActivated(),
        "email" : info.Account.Email(),
        "result" : "ok",
        "username" : info.Account.Username(),
    }, nil
}


func POST__api__user__self(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, NotLoggedInError()
    }
    for fieldName, value := range info.BodyObj {
        switch fieldName {
        case "email":
            oldEmail := info.Account.Email()

            newEmail, ok := value.(string)
            if !ok {
                return nil, BadInputError("Expected string \"email\"")
            }

            // If email hasn't changed, do nothing
            if newEmail == oldEmail {
                break
            }

            err := info.Account.SetEmail(newEmail)
            if err != nil {
                // TODO: finer-grained error reporting
                return nil, InternalServerError("Problem changing email:" + err.Error())
            }

            // send email notifying user of the change.
            skipEmail := false
            skipEmailItf, ok := info.BodyObj["skip-email"]
            if ok {
                skipEmail, ok = skipEmailItf.(bool)
                if !ok {
                    return nil, BadInputError("\"skip-email\" must be boolean").Log()
                }
            }
            if !skipEmail {
                protocol := "http://"
                if info.Config.OptEnableHTTPS() {
                    protocol = "https://"
                }

                activationLink := protocol + info.Config.OptHostname() + 
                    "/mgr/activate.html?username=" + info.Account.Username() + 
                    "&code=" + info.Account.ActivationCode()

                awayMsg := sideEffect.SendEmail()
                awayMsg.AddTo(oldEmail, info.Account.Username())
                awayMsg.SetFrom("no-reply@canopy.link", "Canopy Cloud Service")
                awayMsg.SetReplyTo("no-reply@canopy.link")
                messages.MailMessageEmailChangedAway(awayMsg,
                    info.Account.Username(), 
                    activationLink,
                    protocol + info.Config.OptHostname(),
                    info.Config.OptHostname())

                toMsg := sideEffect.SendEmail()
                toMsg.AddTo(newEmail, info.Account.Username())
                toMsg.SetFrom("no-reply@canopy.link", "Canopy Cloud Service")
                toMsg.SetReplyTo("no-reply@canopy.link")
                messages.MailMessageEmailChangedTo(toMsg,
                    info.Account.Username(), 
                    activationLink,
                    protocol + info.Config.OptHostname(),
                    info.Config.OptHostname())
            }

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

// Delete current account
// Also has side effect of logging the user out, if authenticated with session
// cookie.
func DELETE__api__user__self(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, NotLoggedInError().Log()
    }

    skipEmail := false
    skipEmailItf, ok := info.BodyObj["skip-email"]
    if ok {
        skipEmail, ok = skipEmailItf.(bool)
        if !ok {
            return nil, BadInputError("\"skip-email\" must be boolean").Log()
        }
    }

    // Delete account
    err := info.Conn.DeleteAccount(info.Account.Username())
    if err != nil {
        return nil, InternalServerError("Problem deleting account: " + err.Error()).Log()
    }

    // Send farewell email to the user
    if !skipEmail {
        msg := sideEffect.SendEmail()
        msg.AddTo(info.Account.Email(), info.Account.Username())
        msg.SetFrom("no-reply@canopy.link", "Canopy Cloud Service")
        msg.SetReplyTo("no-reply@canopy.link")
        messages.MailMessageAccountDeleted(msg,
            info.Account.Username(), 
            info.Config.OptHostname(),
        )
    }

    // Log the user out
    sideEffect.Logout()

    return map[string]interface{}{
        "result" : "ok",
    }, nil
}
