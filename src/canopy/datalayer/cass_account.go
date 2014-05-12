package datalayer

import (
    "github.com/gocql/gocql"
    "log"
)

func (dl *CassandraDatalayer) CreateAccount(username string, email string, password string) {
    if err := dl.session.Query(`
            INSERT INTO accounts (username, email, password_hash)
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
    var email string
    if err := dl.session.Query(`
            SELECT email FROM accounts 
            WHERE username = ?
            LIMIT 1
    `, username).Consistency(gocql.One).Scan(&email); err != nil {
            log.Print(err)
            return ""
    }
    return email
}

func (dl *CassandraDatalayer) DeleteAccount(username string) {
    email := dl.GetAccountEmail(username)

    if err := dl.session.Query(`
            DELETE FROM accounts
            WHERE username = ?
    `, username).Exec(); err != nil {
        log.Print(err)
    }

    if err := dl.session.Query(`
            DELETE FROM account_emails
            WHERE email = ?
    `, email).Exec(); err != nil {
        log.Print(err)
    }
}
