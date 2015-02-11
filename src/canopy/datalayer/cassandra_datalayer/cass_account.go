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
package cassandra_datalayer

import (
    "canopy/datalayer"
    "canopy/util/random"
    "code.google.com/p/go.crypto/bcrypt"
    "errors"
    "fmt"
    "time"
    "github.com/gocql/gocql"
)

type CassAccount struct {
    conn *CassConnection
    username string
    email string
    password_hash []byte
    activated bool
    activation_code string
    password_reset_code string
    password_reset_code_expiry time.Time
}

func (account *CassAccount) ActivationCode() string {
    return account.activation_code
}

func (account *CassAccount) Activate(username, code string) error {
    if username != account.Username() {
        return fmt.Errorf("Incorrect username for activation")
    }

    if code != account.ActivationCode() {
        return fmt.Errorf("Incorrect code for activation")
    }

    err := account.conn.session.Query(`
            UPDATE accounts
            SET activated = true
            WHERE username = ?
    `, username).Exec()
    if err != nil {
        return err;
    }

    account.activated = true
    return nil;
}


// Obtain list of devices I have access to.
func (account *CassAccount) Devices() ([]datalayer.Device, error) {
    devices := []datalayer.Device{}
    var deviceId gocql.UUID
    var accessLevel int

    query := account.conn.session.Query(`
            SELECT device_id, access_level FROM device_permissions 
            WHERE username = ?
    `, account.Username()).Consistency(account.conn.dl.readConsistency)
    iter := query.Iter()
    for iter.Scan(&deviceId, &accessLevel) {
        if accessLevel > 0 {
            device, err := account.conn.LookupDevice(deviceId)
            if err != nil {
                iter.Close()
                return []datalayer.Device{}, err
            }
            devices = append(devices, device)
        }
    }
    if err := iter.Close(); err != nil {
        return []datalayer.Device{}, err
    }

    return devices, nil
}

// Obtain specific device, if I have permission.
func (account *CassAccount) Device(id gocql.UUID) (datalayer.Device, error) {
    var accessLevel int

    if err := account.conn.session.Query(`
        SELECT access_level FROM device_permissions
        WHERE username = ? AND device_id = ?
        LIMIT 1
    `, account.Username(), id).Consistency(account.conn.dl.readConsistency).Scan(
        &accessLevel); err != nil {
            return nil, err
    }

    if (accessLevel == datalayer.NoAccess) {
        return nil, errors.New("insufficient permissions ");
    }

    device, err := account.conn.LookupDevice(id)
    if err != nil {
        return nil, err
    }

    return device, nil
}

func (account *CassAccount)Email() string {
    return account.email
}

func (account *CassAccount)GenResetPasswordCode() (string, error) {
    // Generate Password Reset Code
    reset_code, err := random.Base64String(24)
    if err != nil {
        return "", err
    }

    expiry := time.Now().Add(time.Hour*24)
    
    err = account.conn.session.Query(`
            UPDATE accounts
            SET password_reset_code = ?,
                password_reset_code_expiry = ?
            WHERE username = ?
    `, reset_code, expiry, account.Username()).Exec()
    if err != nil {
        return "", err;
    }
    account.password_reset_code = reset_code
    account.password_reset_code_expiry = expiry
    return reset_code, nil
}

func (account *CassAccount) IsActivated() bool {
    return account.activated
}

func (account *CassAccount) ResetPassword(code, newPassword string) error {
    // Verify the code is valid and not expired.
    
    if code == "" || (account.password_reset_code != code) {
        return errors.New("Invalid or expired password reset code");
    }
    if account.password_reset_code_expiry.Before(time.Now()) {
        return errors.New("Invalid or expired password reset code");
    }

    err := account.SetPassword(newPassword)
    if err != nil {
        return err
    }

    pastExpiry := time.Now().Add(-time.Hour*24)

    // Invalidate the code
    err = account.conn.session.Query(`
            UPDATE accounts
            SET password_reset_code = ?,
                password_reset_code_expiry = ?
            WHERE username = ?
    `, "", pastExpiry, account.Username()).Exec()
    if err != nil {
        return err;
    }
    account.password_reset_code = ""
    account.password_reset_code_expiry = pastExpiry
    return nil
}

func (account *CassAccount) SetPassword(password string) error {
    err := validatePassword(password)
    if err != nil {
        return err
    }

    salt := account.conn.dl.cfg.OptPasswordSecretSalt()
    hashCost := account.conn.dl.cfg.OptPasswordHashCost()

    password_hash, err := bcrypt.GenerateFromPassword([]byte(password + salt), int(hashCost))
    if err != nil {
        return err
    }

    err = account.conn.session.Query(`
            UPDATE accounts
            SET password_hash = ?
            WHERE username = ?
    `, password_hash, account.Username()).Exec()
    if err != nil {
        return err;
    }

    account.password_hash = password_hash;
    return nil;
}

func (account *CassAccount)Username() string {
    return account.username
}

func (account* CassAccount)VerifyPassword(password string) bool {
    salt := account.conn.dl.cfg.OptPasswordSecretSalt()
    err := bcrypt.CompareHashAndPassword(account.password_hash, []byte(password + salt))
    return (err == nil)
}
