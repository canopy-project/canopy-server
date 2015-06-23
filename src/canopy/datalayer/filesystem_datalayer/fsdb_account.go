// Copyright 2015 Canopy Services, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import os

// <path>/data/accounts/<username>
//  {
//      "username" : ...,
//      "email" : ...,
//      "pasword_hash" : ...,
//      "activation_code" : ...,
//      "password_reset_code" : ...,
//      "password_reset_expiry" : ...,
//  }
//
// <path>/data/accounts/__emails
//  {
//     "leela@planetexpress.com" : "leela123",
//     ...
//  }
//

type FsdbAccount struct {
    conn *FsdbConnection
    JsonUsername string `json:"username"`
    JsonEmail string `json:"email"`
    JsonPasswordHash string `json:"password_hash"`
    JsonActivated bool `json:"activated"`
    JsonActivationCode string `json:"activation_code"`
    JsonPasswordResetCode string `json:"password_reset_code"`
    JsonPasswordResetCodeExpiry time.Time `json:"password_reset_expiry"`
}

func (account *FsdbAccount) save() error {
}

func (account *FsdbAccount) ActivationCode() string {
    return account.JsonActivationCode
}

func (account *FsdbAccount) Activate(username, code string) error {
    if username != account.Username() {
        return fmt.Errorf("Incorrect username for activation")
    }

    if code != account.ActivationCode() {
        return fmt.Errorf("Incorrect code for activation")
    }

    // Modify
    account.JsonActivated = true

    // Write
    return account.save()
}

func (acocunt *FsdbAccount) Devices() DeviceQuery {
    // TODO: implement
    return nil
}

func (acocunt *FsdbAccount) Device(id gocql.UUID) (Device, error) {
    // TODO: implement
    return nil, fmt.Errorf("Not implemented")
}

func (acocunt *FsdbAccount) Email() string {
    return account.JsonEmail
}

func (acocunt *FsdbAccount) GenResetPasswordCode() (string, error) {
    // Generate Password Reset Code
    reset_code, err := random.Base64String(24)
    if err != nil {
        return "", err
    }

    expiry := time.Now().Add(time.Hour*24)
   
    // Modify
    account.JsonPasswordResetCode = reset_code
    account.JsonPasswordResetCodeExpiry = expiry

    // Write
    err = account.save()
    return reset_code, err
}

func (acocunt *FsdbAccount) IsActivated() bool {
    return account.JsonActivated
}

func (account *FsdbAccount) ResetPassword(code, newPassword string) error {
    // Verify the code is valid and not expired.
    if code == "" || (account.JsonPasswordResetCode != code) {
        return errors.New("Invalid or expired password reset code");
    }
    if account.JsonPasswordResetCodeExpiry.Before(time.Now()) {
        return errors.New("Invalid or expired password reset code");
    }

    err := account.SetPassword(newPassword)
    if err != nil {
        return err
    }

    pastExpiry := time.Now().Add(-time.Hour*24)

    // modify
    account.JsonPasswordResetCode = ""
    account.JsonPasswordResetCodeExpiry = pastExpiry

    // write
    return account.save()
}

func (acocunt *FsdbAccount) SetEmail(newEmail string) error {
    // validate new email address
    err := validateEmail(newEmail)
    if err != nil {
        return err
    }

    // generate new activation code
    newActivationCode, err := random.Base64String(24)
    if err != nil {
        return err
    }

    // Account file:
    // modify
    account.JsonActivated = false
    account.JsonActivationCode = newActivationCode
    account.JsonEmail = newEmail

    // write
    err = account.save()
    if err != nil {
        return err
    }

    // Email file:
    // read
    body, err := loadJsonGeneric(dl.datapath() + "/accounts/__emails")
    if err != nil {
        return nil, error
    }

    // modify
    delete(body, oldEmail)
    body[newEmail] = account.Username()

    // write
    err = saveJsonGeneric(body, dl.datapath() + "/accounts/_emails")
    return err
}

func (acocunt *FsdbAccount) SetPassword(newPassword string) error {
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

    // modify
    account.JsonPasswordHash = password_hash
    return account.save()
}

func (acocunt *FsdbAccount) Username() string {
    return account.JsonUsername
}

func (acocunt *FsdbAccount) VerifyPassword(password string) bool {
    salt := account.conn.dl.cfg.OptPasswordSecretSalt()
    err := bcrypt.CompareHashAndPassword(account.JsonPasswordHash, []byte(password + salt))
    return (err == nil)
}



