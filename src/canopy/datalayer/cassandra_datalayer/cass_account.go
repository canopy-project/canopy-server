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
    "github.com/gocql/gocql"
    "log"
    "errors"
    "code.google.com/p/go.crypto/bcrypt"
)

// Salt is added to passwords.  TODO: don't reveal in source code!
var salt = "aik897sipz0Z*@4:zikp"

 // Computational cost of between 4 and 31.. 14 takes about 1 sec to compute
var hashCost = 10

type CassAccount struct {
    conn *CassConnection
    username string
    email string
    password_hash []byte
}

// Obtain list of devices I have access to.
func (account *CassAccount) Devices() ([]*CassDevice, error) {
    devices := []*CassDevice{}
    var deviceId gocql.UUID
    var accessLevel int

    query := account.conn.session.Query(`
            SELECT device_id, access_level FROM device_permissions 
            WHERE username = ?
    `, account.Username()).Consistency(gocql.One)
    iter := query.Iter()
    for iter.Scan(&deviceId, &accessLevel) {
        if accessLevel > 0 {
            device, err := account.conn.LookupDevice(deviceId)
            if err != nil {
                iter.Close()
                return []*CassDevice{}, err
            }
            devices = append(devices, device)
        }
    }
    if err := iter.Close(); err != nil {
        return []*CassDevice{}, err
    }

    return devices, nil
}

// Obtain specific device, if I have permission.
func (account *CassAccount) Device(id gocql.UUID) (*CassDevice, error) {
    var accessLevel int

    if err := account.conn.session.Query(`
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

    device, err := account.conn.LookupDevice(id)
    if err != nil {
        return nil, err
    }

    return device, nil
}

func (account* CassAccount)Email() string {
    return account.email
}

func (account* CassAccount)Username() string {
    return account.username
}

func (account* CassAccount)VerifyPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword(account.password_hash, []byte(password + salt))
    return (err == nil)
}
