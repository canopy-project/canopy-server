package datalayer

import (
    "github.com/gocql/gocql"
    "log"
    "fmt"
    "code.google.com/p/go.crypto/bcrypt"
)

var salt = "aik897sipz0Z*@4:zikp"
var hashCost = 10 // between 4 and 31.. 14 takes about 1 sec to compute

type CassandraAccount struct {
    username string
    email string
    password_hash []byte
}

type InvalidPasswordError struct{}

func (err InvalidPasswordError) Error() string {
    return fmt.Sprintf("Incorrect password")
}

func (dl *CassandraDatalayer) CreateAccount(username string, email string, password string) (*CassandraAccount, error) {
    password_hash, _ := bcrypt.GenerateFromPassword([]byte(password + salt), hashCost)

    // TODO: transactionize
    if err := dl.session.Query(`
            INSERT INTO accounts (username, email, password_hash)
            VALUES (?, ?, ?)
    `, username, email, password_hash).Exec(); err != nil {
        return nil, err
    }

    if err := dl.session.Query(`
            INSERT INTO account_emails (email, username)
            VALUES (?, ?)
    `, email, username).Exec(); err != nil {
        return nil, err
    }

    return &CassandraAccount{username, email, password_hash}, nil
}

func (dl *CassandraDatalayer)LookupAccount(usernameOrEmail string) (*CassandraAccount, error) {
    var account CassandraAccount

    if err := dl.session.Query(`
            SELECT username, email, password_hash FROM accounts 
            WHERE username = ?
            LIMIT 1
    `, usernameOrEmail).Consistency(gocql.One).Scan(
         &account.username, &account.email, &account.password_hash); err != nil {
            return nil, err
    }
    /* TODO: try email if username not found */
    return &account, nil
}

func (dl *CassandraDatalayer)LookupAccountVerifyPassword(usernameOrEmail string, password string) (*CassandraAccount, error) {
    account, err := dl.LookupAccount(usernameOrEmail)
    if err != nil {
        return nil, err
    }

    verified := account.VerifyPassword(password)
    if (!verified) {
        return nil, &InvalidPasswordError{}
    }

    return account, nil
}

func (account* CassandraAccount)GetUsername() string {
    return account.username
}

func (account* CassandraAccount)GetEmail() string {
    return account.email
}

func (account* CassandraAccount)VerifyPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword(account.password_hash, []byte(password + salt))
    return (err == nil)
}

func (dl *CassandraDatalayer) DeleteAccount(username string) {
    account, _ := dl.LookupAccount(username)
    email := account.GetEmail()

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
