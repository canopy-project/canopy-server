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
    "errors"
    "fmt"
    "github.com/sendgrid/sendgrid-go"
    "time"
)

type CanopySGClient struct {
    sg *sendgrid.SGClient
}

type CanopySGMail struct {
    sgmail *sendgrid.SGMail
}

func NewSendGridMailClient(sendgridUsername, sendgridSecretKey string) (MailClient, error) {
    client := CanopySGClient{}

    if sendgridUsername == "" {
        return nil, fmt.Errorf("sendgrid-username must be set")
    }

    if sendgridSecretKey == "" {
        return nil, fmt.Errorf("sendgrid-secret-key must be set")
    }

    client.sg = sendgrid.NewSendGridClient(sendgridUsername, sendgridSecretKey)
    return &client, nil
}

func newSendGridMessage() (MailMessage) {
    mail := CanopySGMail{}
    mail.sgmail = sendgrid.NewMail()
    return &mail
}

func (*CanopySGClient) NewMail() MailMessage {
    return newSendGridMessage()
}
func (client *CanopySGClient) Send(m MailMessage) error {
    mail, ok := m.(*CanopySGMail)
    if !ok {
        return errors.New("Message was not constructed with CanopySGClient")
    }
    err := client.sg.Send(mail.sgmail)
    return err
}

func (mail *CanopySGMail) AddTo(email string, name string) error {
    err := mail.sgmail.AddTo(email)
    if err != nil {
        return err
    }
    if name != "" {
        mail.sgmail.AddToName(name)
    }
    return nil
}

func (mail *CanopySGMail) AddTos(emails []string, names []string) error {
    err := mail.sgmail.AddTos(emails)
    if err != nil {
        return err
    }
    mail.sgmail.AddToNames(names)
    return nil
}

func (mail *CanopySGMail) SetSubject(subject string) {
    mail.sgmail.SetSubject(subject)
}

func (mail *CanopySGMail) SetText(text string) {
    mail.sgmail.SetText(text)
}

func (mail *CanopySGMail) SetHTML(html string) {
    mail.sgmail.SetHTML(html)
}

func (mail *CanopySGMail) SetFrom(email string, name string) error {
    err := mail.sgmail.SetFrom(email)
    if err != nil {
        return err
    }
    mail.sgmail.SetFromName(name)
    return nil
}

func (mail *CanopySGMail) SetReplyTo(email string) error {
    return mail.sgmail.SetReplyTo(email)
}

func (mail *CanopySGMail) SetDate(date time.Time) error {
    return errors.New("Not implemented")
}
