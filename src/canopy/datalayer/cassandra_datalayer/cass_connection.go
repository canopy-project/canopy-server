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

import(
    "canopy/canolog"
    "canopy/datalayer"
    "canopy/sddl"
    "github.com/gocql/gocql"
    "code.google.com/p/go.crypto/bcrypt"
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

func (conn *CassConnection) CreateAccount(username, email, password string) (datalayer.Account, error) {
    password_hash, _ := bcrypt.GenerateFromPassword([]byte(password + salt), hashCost)

    // TODO: transactionize
    if err := conn.session.Query(`
            INSERT INTO accounts (username, email, password_hash)
            VALUES (?, ?, ?)
    `, username, email, password_hash).Exec(); err != nil {
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

    return &CassAccount{conn, username, email, password_hash}, nil
}

func (conn *CassConnection) CreateDevice(name string, uuid *gocql.UUID, publicAccessLevel datalayer.AccessLevel) (datalayer.Device, error) {
    // TODO: validate parameters 
    var id gocql.UUID
    if uuid == nil {
        id = gocql.TimeUUID()
    } else {
        id = *uuid
    }
    
    err := conn.session.Query(`
            INSERT INTO devices (device_id, friendly_name, public_access_level)
            VALUES (?, ?, ?)
    `, id, name, publicAccessLevel).Exec()
    if err != nil {
        canolog.Error("Error creating device:", err)
        return nil, err
    }
    return &CassDevice{
        conn: conn,
        deviceId: id,
        name: name,
        class: nil,         // class gets initialized during first report
        classString: "",
        publicAccessLevel: publicAccessLevel,
    }, nil
}

func (conn *CassConnection) LookupOrCreateDevice(deviceId gocql.UUID, publicAccessLevel datalayer.AccessLevel) (datalayer.Device, error) {
    // TODO: improve this implementation.
    // Fix race conditions?
    // Fix error paths?
    
    device, err := conn.LookupDevice(deviceId)
    if device != nil {
        canolog.Info("LookupOrCreateDevice - device ", deviceId, " found")
        return device, nil
    }

    device, err = conn.CreateDevice("AnonDevice", &deviceId, publicAccessLevel)
    if err != nil {
        canolog.Info("LookupOrCreateDevice - device ", deviceId, "error")
    }
    canolog.Info("LookupOrCreateDevice - device ", deviceId, " created")
    return device, err
}

func (conn *CassConnection) DeleteAccount(username string) {
    account, _ := conn.LookupAccount(username)
    email := account.Email()

    if err := conn.session.Query(`
            DELETE FROM accounts
            WHERE username = ?
    `, username).Exec(); err != nil {
        canolog.Error("Error deleting account", err)
    }

    if err := conn.session.Query(`
            DELETE FROM account_emails
            WHERE email = ?
    `, email).Exec(); err != nil {
        canolog.Error("Error deleting account email", err)
    }
}

func (conn *CassConnection) LookupAccount(usernameOrEmail string) (datalayer.Account, error) {
    var account CassAccount

    if err := conn.session.Query(`
            SELECT username, email, password_hash FROM accounts 
            WHERE username = ?
            LIMIT 1
    `, usernameOrEmail).Consistency(gocql.One).Scan(
         &account.username, &account.email, &account.password_hash); err != nil {
            canolog.Error("Error looking up account", err)
            return nil, err
    }
    /* TODO: try email if username not found */
    account.conn = conn
    return &account, nil
}

func (conn *CassConnection)LookupAccountVerifyPassword(usernameOrEmail string, password string) (datalayer.Account, error) {
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

func (conn *CassConnection) LookupDevice(deviceId gocql.UUID) (datalayer.Device, error) {
    var device CassDevice

    device.deviceId = deviceId
    device.conn = conn

    err := conn.session.Query(`
        SELECT friendly_name, sddl
        FROM devices
        WHERE device_id = ?
        LIMIT 1`, deviceId).Consistency(gocql.One).Scan(
            &device.name,
            &device.classString)
    if err != nil {
        return nil, err
    }

    if device.classString != "" {
        device.class, err = sddl.ParseClassString("anonymous", device.classString)
        if err != nil {
            canolog.Error("Error parsing class string for device: ", device.classString, err)
            return nil, err
        }
    }

    return &device, nil
}

func (conn *CassConnection) LookupDeviceByStringID(id string) (datalayer.Device, error) {
    deviceId, err := gocql.ParseUUID(id)
    if err != nil {
        canolog.Error(err)
        return nil, err
    }
    return conn.LookupDevice(deviceId)
}

