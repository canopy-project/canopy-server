/*
 * Copright 2014-2015 Canopy Services, Inc.
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

import(
    "canopy/canolog"
    "canopy/datalayer"
    "canopy/sddl"
    "canopy/util/random"
    "github.com/gocql/gocql"
    "code.google.com/p/go.crypto/bcrypt"
    "regexp"
    "strings"
    "time"
)

type CassConnection struct {
    dl *CassDatalayer
    session *gocql.Session
}

// Use with care.  Erases all sensor data.
func (conn *CassConnection) ClearSensorData() {
    tables := []string{
        "propval_int",
        "propval_float",
        "propval_double",
        "propval_timestamp",
        "propval_boolean",
        "propval_void",
        "propval_string",
    }
    for _, table := range tables {
        err := conn.session.Query(`TRUNCATE ` + table).Exec();
        if (err != nil) {
            canolog.Error("Error truncating ", table, ":", err)
        }
    }
}

func (conn *CassConnection) Close() {
    conn.session.Close()
}

func validateUsername(username string) error {
    if username == "leela" {
        return datalayer.NewValidationError("Username reserved")
    }
    if len(username) < 5 {
        return datalayer.NewValidationError("Username too short")
    }
    if len(username) > 24 {
        return datalayer.NewValidationError("Username too long")
    }
    matched, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_]+$", username)
    if !matched || err != nil {
        return datalayer.NewValidationError("Invalid username")
    }

    return nil
}

func validatePassword(password string) error {
    if len(password) < 6 {
        return datalayer.NewValidationError("Password too short")
    }
    if len(password) > 120 {
        return datalayer.NewValidationError("Password too long")
    }
    return nil
}

var emailPattern = regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")
func validateEmail(email string) error {
    if !emailPattern.MatchString(email) {
        return datalayer.NewValidationError("Invalid email address")
    }
    return nil
}

func (conn *CassConnection) CreateAccount(
        username, 
        email, 
        password string) (datalayer.Account, error) {

    salt := conn.dl.cfg.OptPasswordSecretSalt()
    hashCost := conn.dl.cfg.OptPasswordHashCost()

    password_hash, _ := bcrypt.GenerateFromPassword(
            []byte(password + salt), int(hashCost))

    err := validateUsername(username)
    if err != nil {
        return nil, err
    }

    err = validateEmail(email)
    if err != nil {
        return nil, err
    }

    err = validatePassword(password)
    if err != nil {
        return nil, err
    }

    activation_code, err := random.Base64String(24)
    if err != nil {
        return nil, err
    }

    now := time.Now()

    // TODO: transactionize
    if err := conn.session.Query(`
            INSERT INTO accounts (
                username, 
                email, 
                password_hash, 
                activated, 
                activation_code,
                password_reset_code,
                password_reset_code_expiry)
            VALUES (?, ?, ?, ?, ?, ?, ?)
    `, username, email, password_hash, false, activation_code, "", now).Exec(); err != nil {
        canolog.Error("Error creating account:", err)
        return nil, err
    }

    if err := conn.session.Query(`
            INSERT INTO account_emails (email, username)
            VALUES (?, ?)
    `, email, username).Exec(); err != nil {
        canolog.Error("Error setting account email:", err)
        return nil, err
    }

    return &CassAccount{conn, username, email, password_hash, false, activation_code, "", now}, nil
}

func (conn *CassConnection) CreateDevice(
        name string, 
        uuid *gocql.UUID, 
        secretKey string, 
        publicAccessLevel datalayer.AccessLevel) (datalayer.Device, error) {
    // TODO: validate parameters 
    var id gocql.UUID
    var err error

    if uuid == nil {
        id, err = gocql.RandomUUID()
        if err != nil {
            return nil, err
        }
    } else {
        id = *uuid
    }
    
    if secretKey == "" {
        secretKey, err = random.Base64String(24)
        if err != nil {
            return nil, err
        }
    }
    
    err = conn.session.Query(`
            INSERT INTO devices (device_id, secret_key, friendly_name, public_access_level)
            VALUES (?, ?, ?, ?)
    `, id, secretKey, name, publicAccessLevel).Exec()
    if err != nil {
        canolog.Error("Error creating device:", err)
        return nil, err
    }
    return &CassDevice{
        conn: conn,
        deviceId: id,
        secretKey: secretKey,
        name: name,
        doc: sddl.Sys.NewEmptyDocument(),
        docString: "",
        publicAccessLevel: publicAccessLevel,
        locationNote: "",
        wsConnected: false,
    }, nil
}

func (conn *CassConnection) DeleteAccount(username string) error {
    // TODO: We should archive the account, not actually delete it.
    // TODO: If we are really deleting it, then we need to also cleanup
    // all the other data (permissions, orphanded devices, etc).
    account, err := conn.LookupAccount(username)
    if err != nil {
        canolog.Error("Error looking up account for deletion: ", err)
        return err
    }

    email := account.Email()

    // TODO: Transactionize.  This might be done by adding a txn state field to
    // the table.

    err = conn.session.Query(`
            DELETE FROM device_group
            WHERE username = ?
    `, username).Exec()
    if err != nil {
        canolog.Error("Error deleting account's device groups", err)
        return err
    }

    err = conn.session.Query(`
            DELETE FROM device_permissions
            WHERE username = ?
    `, username).Exec()
    if err != nil {
        canolog.Error("Error deleting account's permission", err)
        return err
    }

    err = conn.session.Query(`
            DELETE FROM account_emails
            WHERE email = ?
    `, email).Exec()
    if err != nil {
        canolog.Error("Error deleting account email", err)
        return err
    }

    err = conn.session.Query(`
            DELETE FROM accounts
            WHERE username = ?
    `, username).Exec()
    if err != nil {
        canolog.Error("Error deleting account", err)
        return err
    }

    return nil
}

func (conn *CassConnection)DeleteDevice(deviceId gocql.UUID) error {
    // TODO: Should we archive the device, not actually delete it?
    device, err := conn.LookupDevice(deviceId)
    if err != nil {
        canolog.Error("Error deleting device", err)
        return err
    }

    err = conn.session.Query(`
            DELETE FROM devices
            WHERE device_id = ?
    `, device.ID()).Exec()
    if err != nil {
        canolog.Error("Error deleting from devices table", err)
        return err
    }

    // TODO: How to cleanup device_permissions?

    // TODO: transactionize
    // TODO: Cleanup cloud variable data
    return nil
}

func (conn *CassConnection) LookupAccount(
        usernameOrEmail string) (datalayer.Account, error) {
    var account CassAccount
    var username string

    canolog.Info("Looking up account: ", usernameOrEmail)

    if strings.Contains(usernameOrEmail, "@") {
        canolog.Info("It is an email address")
        // email address provided.  Lookup username based on email
        err := conn.session.Query(`
                SELECT email, username FROM account_emails
                WHERE email = ?
                LIMIT 1
        `, usernameOrEmail).Consistency(gocql.One).Scan(
             &account.email, &username);
        
        if (err != nil) {
            canolog.Error("Error looking up account", err)
            return nil, err
        }
    } else {
        canolog.Info("It is not an email address")
        username = usernameOrEmail
    }

    canolog.Info("fetching info for: ", username)
    // Lookup account info based on username
    err := conn.session.Query(`
            SELECT 
                username, 
                email, 
                password_hash, 
                activated, 
                activation_code, 
                password_reset_code, 
                password_reset_code_expiry 
            FROM accounts 
            WHERE username = ?
            LIMIT 1
    `, username).Consistency(gocql.One).Scan(
         &account.username, 
         &account.email, 
         &account.password_hash, 
         &account.activated, 
         &account.activation_code,
         &account.password_reset_code,
         &account.password_reset_code_expiry)
    
    if (err != nil) {
        canolog.Error("Error looking up account", err)
        return nil, err
    }

    canolog.Info("Success")
    account.conn = conn
    return &account, nil
}

func (conn *CassConnection) LookupAccountVerifyPassword(
        usernameOrEmail string, 
        password string) (datalayer.Account, error) {
    account, err := conn.LookupAccount(usernameOrEmail)
    if err != nil {
        return nil, err
    }

    verified := account.VerifyPassword(password)
    if (!verified) {
        canolog.Info("Incorrect password for ", usernameOrEmail)
        return nil, datalayer.InvalidPasswordError
    }

    return account, nil
}

func (conn *CassConnection) LookupDevice(
        deviceId gocql.UUID) (datalayer.Device, error) {
    var device CassDevice

    device.deviceId = deviceId
    device.conn = conn
    var last_seen time.Time
    var ws_connected bool

    err := conn.session.Query(`
        SELECT friendly_name, location_note, secret_key, sddl, last_seen, ws_connected
        FROM devices
        WHERE device_id = ?
        LIMIT 1`, deviceId).Consistency(gocql.One).Scan(
            &device.name,
            &device.locationNote,
            &device.secretKey,
            &device.docString,
            &last_seen,
            &ws_connected)
    if err != nil {
        canolog.Error(err)
        return nil, err
    }

    // This scan returns Jan 1, 1970 UTC if last_seen is NULL.
    if last_seen.Before(time.Unix(1, 0)) {
        device.last_seen = nil
    } else {
        device.last_seen = &last_seen
    }

    device.wsConnected = ws_connected

    if device.docString != "" {
        device.doc, err = sddl.Sys.ParseDocumentString(device.docString)
        if err != nil {
            canolog.Error("Error parsing class string for device: ", device.docString, err)
            return nil, err
        }
    } else {
        device.doc = sddl.Sys.NewEmptyDocument()
    }

    return &device, nil
}
func (conn *CassConnection) LookupDeviceVerifySecretKey(
        deviceId gocql.UUID, 
        secret string) (datalayer.Device, error) {

    device, err := conn.LookupDevice(deviceId)
    if err != nil {
        return nil, err
    }

    if device.SecretKey() != secret {
        canolog.Error("Invalid secret key")
        return nil, datalayer.InvalidPasswordError
    }

    return device, nil
}

func (conn *CassConnection) LookupDeviceByStringID(
        id string) (datalayer.Device, error) {

    deviceId, err := gocql.ParseUUID(id)
    if err != nil {
        canolog.Error(err)
        return nil, err
    }
    return conn.LookupDevice(deviceId)
}

func (conn *CassConnection) LookupDeviceByStringIDVerifySecretKey(
        id, 
        secret string) (datalayer.Device, error) {

    deviceId, err := gocql.ParseUUID(id)
    if err != nil {
        canolog.Error(err)
        return nil, err
    }
    return conn.LookupDeviceVerifySecretKey(deviceId, secret)
}

func (conn *CassConnection) PigeonSystem() datalayer.PigeonSystem {
    return &CassPigeonSystem{conn}
}
