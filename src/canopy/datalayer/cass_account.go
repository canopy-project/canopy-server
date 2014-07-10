package datalayer

import (
    "github.com/gocql/gocql"
    "log"
    "errors"
    "code.google.com/p/go.crypto/bcrypt"
)

var salt = "aik897sipz0Z*@4:zikp"
var hashCost = 10 // between 4 and 31.. 14 takes about 1 sec to compute

type CassandraAccount struct {
    dl *CassandraDatalayer
    username string
    email string
    password_hash []byte
}

var InvalidPasswordError = errors.New("Incorrect password")

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

    return &CassandraAccount{dl, username, email, password_hash}, nil
}

func (dl *CassandraDatalayer) DeleteAccount(username string) {
    account, _ := dl.LookupAccount(username)
    email := account.Email()

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
    account.dl = dl
    return &account, nil
}

func (dl *CassandraDatalayer)LookupAccountVerifyPassword(usernameOrEmail string, password string) (*CassandraAccount, error) {
    account, err := dl.LookupAccount(usernameOrEmail)
    if err != nil {
        return nil, err
    }

    verified := account.VerifyPassword(password)
    if (!verified) {
        return nil, InvalidPasswordError
    }

    return account, nil
}

func (account* CassandraAccount)Username() string {
    return account.username
}

func (account* CassandraAccount)Email() string {
    return account.email
}

func (account* CassandraAccount)VerifyPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword(account.password_hash, []byte(password + salt))
    return (err == nil)
}

/*
 * Obtain list of devices I have access to 
 */
func (account *CassandraAccount)GetDevices() ([]*CassandraDevice, error) {
    devices := []*CassandraDevice{}
    var deviceId gocql.UUID
    var accessLevel int

    query := account.dl.session.Query(`
            SELECT device_id, access_level FROM device_permissions 
            WHERE username = ?
    `, account.Username()).Consistency(gocql.One)
    iter := query.Iter()
    for iter.Scan(&deviceId, &accessLevel) {
        if accessLevel > 0 {
            /* TODO: can we do another query inside an iterator? */
            device, err := account.dl.LookupDevice(deviceId)
            if err != nil {
                iter.Close()
                return []*CassandraDevice{}, err
            }
            devices = append(devices, device)
        }
    }
    if err := iter.Close(); err != nil {
        return []*CassandraDevice{}, err
    }

    return devices, nil
}

/*
 * Obtain specific device, if I have permission.
 */
func (account *CassandraAccount)GetDeviceById(deviceId gocql.UUID) (*CassandraDevice, error) {
    var accessLevel int

    if err := account.dl.session.Query(`
        SELECT access_level FROM device_permissions
        WHERE username = ? AND device_id = ?
        LIMIT 1
    `, account.Username(), deviceId).Consistency(gocql.One).Scan(
        &accessLevel); err != nil {
            return nil, err
    }

    if (accessLevel == NoAccess) {
        return nil, errors.New("insufficient permissions ");
    }

    device, err := account.dl.LookupDevice(deviceId)
    if err != nil {
        return nil, err
    }

    return device, nil
}
