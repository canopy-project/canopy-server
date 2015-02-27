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
    "canopy/jobqueue"
    "canopy/mail"
)

type RestSideEffects struct {
    setCookies map[string]string
    clearCookies []string
    sendEmails []mail.MailMessage
    mailer mail.MailClient
}

// Causes a cookie to be set as a side-effect during REST endpoint handling.
// This does not actually set the cookie.
//
// When Perform() is called, it adds to the Pigeon response: 
//      "set-cookies" : {
//          key0: value0,
//          key1: value1,
//          ...
//      }
//
// The job issuer then must actually set the cookies.
func (sideEffect *RestSideEffects) SetCookie(key, value string) {
    sideEffect.setCookies[key] = value
}

// Causes a cookie to be cleared as a side-effect during REST endpoint
// handling.  This does not actually clear the cookie.
//
// When Perform() is called, it adds to the Pigeon response: 
//      "clear-cookies" : [ key0, key1, ... ]
//
// The job issuer then must actually set the cookies.
func (sideEffect *RestSideEffects) ClearCookie(key string) {
    sideEffect.clearCookies = append(sideEffect.clearCookies, key)
}

// Causes an email to be sent as a side-effect during REST endpoint handling.
// This does not actually send the cookie.
//
// When Perform() is called, the email will be sent.
// 
// Returns a new mail.MailMessage object that the caller must use to compose
// the message.
func (sideEffect *RestSideEffects) SendEmail() mail.MailMessage {
    msg := sideEffect.mailer.NewMail();
    sideEffect.sendEmails = append(sideEffect.sendEmails, msg)
    return msg
}

// Carries out the side-effect actions.
// Specifically:
//
//  1) Sends emails
//  2) Appends "set-cookies" and "clear-cookies" to the response object, as
//  appropriate.
func (sideEffect *RestSideEffects) Perform(req jobqueue.Request, resp jobqueue.Response) error {
    if len(sideEffect.setCookies) > 0 {
        resp.AppendToBody("set-cookies", sideEffect.setCookies)
    }
    if len(sideEffect.clearCookies) > 0 {
        resp.AppendToBody("clear-cookies", sideEffect.clearCookies)
    }
    for _, email := range sideEffect.sendEmails {
        err := sideEffect.mailer.Send(email)
        if err != nil {
            // Log the error, but do not affect the HTTP response
            canolog.Error("Error sending email: ", err.Error())
        }
    }
    return nil
}

// Creates a new RestSideEffects object.
func NewRestSideEffects(mailer mail.MailClient) *RestSideEffects {
    return &RestSideEffects{
        mailer: mailer,
        setCookies: map[string]string{},
        clearCookies: []string{},
        sendEmails: []mail.MailMessage{},
    }
}
