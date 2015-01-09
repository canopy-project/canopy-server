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
package mail

import (
    "canopy/config"
    "fmt"
    "time"
)

type MailClient interface {
    NewMail() MailMessage
    Send(m MailMessage) error
}

type MailMessage interface {
    AddTo(email string, name string) error
    AddTos(emails []string, names []string) error
    SetSubject(subject string)
    SetText(text string)
    SetHTML(html string)
    SetFrom(email string, name string) error
    SetReplyTo(email string) error
    SetDate(date time.Time) error
}

func NewMailClient(cfg config.Config) (MailClient, error) {
    switch cfg.OptEmailService() {
    case "none":
        return NewNoOpMailClient()
    case "sendgrid":
        username := cfg.OptSendgridUsername()
        secret := cfg.OptSendgridSecretKey()
        return NewSendGridMailClient(username, secret)
    default:
        return nil, fmt.Errorf("Unsupported mail service: %s", cfg.OptEmailService())
    }
}
