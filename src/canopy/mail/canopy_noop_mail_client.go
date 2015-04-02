/*
 * Copright 2014-2015 Canopy Services, Inc.
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
    "canopy/canolog"
    "time"
)

type CanopyNoOpMailClient struct {
}

type CanopyNoOpMail struct {
}

func NewNoOpMailClient() (MailClient, error) {
    return &CanopyNoOpMailClient{}, nil
}

func (*CanopyNoOpMailClient) NewMail() MailMessage {
    return &CanopyNoOpMail{}
}
func (client *CanopyNoOpMailClient) Send(m MailMessage) error {
    canolog.Info("Noop Mail Client: not sending message")
    return nil
}

func (mail *CanopyNoOpMail) AddTo(email string, name string) error {
    return nil
}

func (mail *CanopyNoOpMail) AddTos(emails []string, names []string) error {
    return nil
}

func (mail *CanopyNoOpMail) SetSubject(subject string) {
    return
}

func (mail *CanopyNoOpMail) SetText(text string) {
    return
}

func (mail *CanopyNoOpMail) SetHTML(html string) {
    return
}

func (mail *CanopyNoOpMail) SetFrom(email string, name string) error {
    return nil
}

func (mail *CanopyNoOpMail) SetReplyTo(email string) error {
    return nil
}

func (mail *CanopyNoOpMail) SetDate(date time.Time) error {
    return nil
}
