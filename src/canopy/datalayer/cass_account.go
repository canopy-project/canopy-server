package datalayer

import (
    "log"
)

func (dl *CassandraDatalayer) CreateAccount(username string, email string, password string) {
    if err := dl.session.Query(`
            INSERT INTO account (username, email, password_hash)
            VALUES (?, ?, ?)
    `, username, email, password).Exec(); err != nil {
        log.Print(err)
    }

    if err := dl.session.Query(`
            INSERT INTO account_emails (email, username)
            VALUES (?, ?)
    `, email, username).Exec(); err != nil {
        log.Print(err)
    }
}

func (dl *CassandraDatalayer) GetAccountEmail(username string) string {
    return "dunno"
}

func (dl *CassandraDatalayer) DeleteAccount(username string) {
}
