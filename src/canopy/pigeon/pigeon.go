package pigeon

import (
    "errors"
    "time"
)

/*
 * The "pigeon" package is a message-passing system for canopy.
 *
 * It is used to forward control instructions received over HTTP to the
 * appropriate go thread containing the websocket connection for the
 * appropriate device.
 *
 * For now, it only functions locally, but eventually it will work across
 * servers.
 *
 * TODO: fix all the race conditions
 * TODO: switch to buffered
 * TODO: expose 
 */

type PigeonSystem struct {
    mailboxes map[string]*PigeonMailbox
}

type PigeonMailbox struct {
    ch chan *PigeonMessage
    id string
    sys *PigeonSystem
}

type PigeonMessage struct {
    Data map[string]interface{}
}

func InitPigeonSystem() *PigeonSystem {
    return &PigeonSystem{mailboxes: map[string]*PigeonMailbox{}}
}

func (pigeon *PigeonSystem)CreateMailbox(mailboxId string) (*PigeonMailbox) {
    mailbox := PigeonMailbox{make(chan *PigeonMessage), mailboxId, pigeon}
    pigeon.mailboxes[mailboxId] = &mailbox;
    return &mailbox
}

func (pigeon *PigeonSystem)Mailbox(mailboxId string) (*PigeonMailbox) {
    return pigeon.mailboxes[mailboxId];
}

func (pigeon *PigeonSystem)SendMessage(mailboxId string, msg *PigeonMessage, timeout time.Duration) error{
    mailbox := pigeon.mailboxes[mailboxId]
    if mailbox != nil {
        select {
            case mailbox.ch <- msg:
                // message transferred
                return nil
            case <- time.After(timeout):
                return errors.New("SendMessage timed out")
        }
    }
    return errors.New("Mailbox not found")
}

func (mailbox *PigeonMailbox)RecieveMessage(timeout time.Duration) (*PigeonMessage, error) {
    select {
        case msg := <- mailbox.ch:
            return msg, nil
        case <- time.After(timeout):
            return nil, errors.New("ReceiveMessage timed out")
    }
}

func (mailbox *PigeonMailbox)Close() {
    delete(mailbox.sys.mailboxes, mailbox.id)
}
