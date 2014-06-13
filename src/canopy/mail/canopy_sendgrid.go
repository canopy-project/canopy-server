package mail

import (
    "errors"
    "github.com/sendgrid/sendgrid-go"
    "os"
    "time"
)

type CanopySGClient struct {
    sg *sendgrid.SGClient
}

type CanopySGMail struct {
    sgmail *sendgrid.SGMail
}

func NewSendGridMailClient() (MailClient, error) {
    client := CanopySGClient{}
    username := os.Getenv("SENDGRID_USERNAME")
    if username == "" {
        return nil, errors.New("You must set environment variable SENDGRID_USERNAME")
    }

    secret := os.Getenv("SENDGRID_SECRET_KEY")
    if secret == "" {
        return nil, errors.New("You must set environment variable SENDGRID_SECRET_KEY")
    }
    client.sg = sendgrid.NewSendGridClient(username, secret)
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
