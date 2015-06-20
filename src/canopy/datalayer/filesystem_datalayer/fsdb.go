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

//
// Filesystem database:
//
// <path>/data/accounts/<username>
// <path>/data/devices/<id>

type FsdbDatalayer struct {
    cfg config.Config
    path string // i.e. "/var/canopy/db"
}

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

func (dl *FsdbDatalayer) Connect() datalayer.Connection, error {
    return &FsdbConnection{}, nil
}

func (dl *FsdbDatalayer) datapath() string {
    // We store everything in the (<path> + "/data") directory.  This is mostly
    // for damage control if <path> is misconfigured.  This way, we'll never,
    // for example "rm -rf /" wiping the whole file system.
    return path + "/data"
}

func (dl *FsdbDatalayer) EraseDb() error {
    os.RemoveAll(dl.datapath())
}

func (dl *FsdbDatalayer) PrepDb() error {
    return os.MkdirAll(dl.datapath(), 0777)
}

func (dl *FsdbDatalayer) MigrateDb() error {
    return fmt.Errorf("Not implemented")
}


func (conn *FsdbConnection) ClearSensorData() error {
    // TODO: implement
}

func (conn *FsdbConnection) Close() {
    // noop
}

func (conn *FsdbConnection) DeleteAccount(username string) {
    return os.Remove(dl.datapath() + "/accounts/" + username)
}

func (conn *FsdbConnection) DeleteDevice(deviceId gocql.UUID) {
    return os.Remove(dl.datapath() + "/devices/" + deviceId.ToString())
}

func (conn *FsdbConnection) LookupAccount(usernameOrEmail string) (Account, error) {
    // TODO: handle email addresses
    file, err := os.Open(dl.datapath()  + "/accounts/" + usernameOrEmail)
    if err != nil {
        return nil, err
    }

    // Read file and parse JSON
    account = &FsdbAccount{conn: conn}
    err = json.NewDecoder(file).Decode(account)
    if err != nil {
        return err
    }

    return account
}

func (account *FsdbAccount) ActivationCode() string {
    return account.JsonActivationCode
}

func (account *FsdbAccount) Activate(username, code string) error {
    return fmt.Errorf("Not implemented")
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
    return "", fmt.Errorf("Not implemented")
}

func (acocunt *FsdbAccount) IsActivated() bool {
    return account.JsonActivated
}

func (acocunt *FsdbAccount) ResetPassword(code, newPassword string) error {
    return fmt.Errorf("Not implemented")
}

func (acocunt *FsdbAccount) SetEmail(newEmail string) error {
    return fmt.Errorf("Not implemented")
}

func (acocunt *FsdbAccount) SetPassword(newPassword string) error {
    return fmt.Errorf("Not implemented")
}

func (acocunt *FsdbAccount) Username() string {
    return account.JsonUsername
}

func (acocunt *FsdbAccount) VerifyPassword(password string) bool {
    // TODO: implement
    return false
}

func NewDatalayer(cfg config.Config) datalayer.Datalayer {
    return &FsdbDatalayer{cfg: cfg, path: "/var/canopy/db"}
}
