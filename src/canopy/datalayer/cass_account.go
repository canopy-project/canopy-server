package datalayer

import (
    "github.com/gocql/gocql"
    "log"
    "code.google.com/p/go.crypto/bcrypt"
)

var salt = "aik897sipz0Z*@4:zikp"
var hashCost = 10 // between 4 and 31.. 14 takes about 1 sec to compute

func (dl *CassandraDatalayer) CreateAccount(username string, email string, password string) {
    hashed_password, _ := bcrypt.GenerateFromPassword([]byte(password + salt), 14)

    if err := dl.session.Query(`
            INSERT INTO accounts (username, email, password_hash)
            VALUES (?, ?, ?)
    `, username, email, hashed_password).Exec(); err != nil {
        log.Print(err)
    }

    if err := dl.session.Query(`
            INSERT INTO account_emails (email, username)
            VALUES (?, ?)
    `, email, username).Exec(); err != nil {
        log.Print(err)
    }
}

func (dl *CassandraDatalayer) VerifyAccountPassword(username string, password string) bool {
    var hashed_password []byte
    if err := dl.session.Query(`
            SELECT password_hash FROM accounts 
            WHERE username = ?
            LIMIT 1
    `, username).Consistency(gocql.One).Scan(&hashed_password); err != nil {
            log.Print(err)
            return false
    }

    err := bcrypt.CompareHashAndPassword(hashed_password, []byte(password + salt))
    return (err == nil)
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
