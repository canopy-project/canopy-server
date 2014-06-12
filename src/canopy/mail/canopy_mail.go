package mail

import (
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

func NewDefaultMailClient() (MailClient, error) {
    return NewSendGridMailClient()
}

