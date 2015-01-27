// Copyright 2015 SimpleThings, Inc.
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
    "canopy/canolog"
    "canopy/mail/messages"
    "canopy/rest/adapter"
    "canopy/rest/rest_errors"
    "net/http"
)

// This endpoint is used for two purposes:
//
// Purpose 1) Send Reset Password Request.  If only "username" field is
// supplied then an email is sent to that user with a link for resetting
// password.  Request:
// {
//      "username" : <USERNAME_OR_EMAIL>
// }
//
// Purpose 2) Setting new password.  If "username",  "password", and "code" are
// set, then the user's password is reset to the provided password.  Request:
// {
//      "username" : <USERNAME_OR_EMAIL>,
//      "password" : <NEW_PASSWORD>,
//      "code" : <PASSWORD_RESET_CODE>,
// }
func POST_reset_password(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    usernameOrEmail, ok := info.BodyObj["username"].(string)
    if !ok {
        return nil, rest_errors.NewBadInputError("String \"username\" expected")
    }

    account, err := info.Conn.LookupAccount(usernameOrEmail)
    if err != nil {
        // See
        // http://stackoverflow.com/questions/2878990/is-there-a-security-reason-not-to-reveal-the-existence-of-a-user-id

        // Bottom line: For security we should not reveal whether or not a
        // username or password exists, so we simply report that the
        // confirmation email has been sent.
        return map[string]interface{} {
            "result" : "ok",
        }, nil
    }

    password, ok := info.BodyObj["password"].(string)
    if ok {
        // Set new password using code (Purpose 2 above).
        code, ok := info.BodyObj["code"].(string)
        if !ok {
            return nil, rest_errors.NewBadInputError("String \"code\" expected")
        }

        err = account.ResetPassword(code, password)
        if err != nil {
            // TODO: Report InternalServerError different from InvalidCode
            return nil, rest_errors.NewBadInputError("Unable to reset password: " + err.Error())
        }
    } else {
        // Send Reset Password Request (Purpose 1 above)
        canolog.Trace("Sending password reset email")

        protocol := "http://"
        if info.Config.OptEnableHTTPS() {
            protocol = "https://"
        }

        code, err := account.GenResetPasswordCode()
        if (err != nil) {
            return nil, rest_errors.NewInternalServerError("Problem resetting password: " + err.Error())
        }

        activationLink := protocol + info.Config.OptHostname() + 
                "/mgr/reset_password.html?username=" + account.Username() + 
                "&code=" + code

        msg := info.Mailer.NewMail();
        msg.AddTo(account.Email(), account.Username())
        msg.SetFrom("no-reply@canopy.link", "Canopy Cloud Service")
        msg.SetReplyTo("no-reply@canopy.link")
        messages.MailMessageResetPassword(msg,
            account.Username(), 
            activationLink,
            protocol + info.Config.OptHostname(),
            info.Config.OptHostname(),
        )
        err = info.Mailer.Send(msg)
        if (err != nil) {
            return nil, rest_errors.NewInternalServerError("Problem sending mail")
        }
    }

    return map[string]interface{} {
        "result" : "ok",
    }, nil
}
