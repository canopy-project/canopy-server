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

type FsdbDevice struct {
    conn *FsdbConnection
    doc sddl.Document
    JsonID string `json:"id"`
    JsonLastSeen time.Time `json:"last_seen"`
    JsonLocationNote string `json:"location_note"`
    JsonName string `json:"name"`
    JsonPublicAccessLevel int `json:"public_access_level"`
    JsonSecretKey string `json:"secret_key"`
    JsonWSConnected bool `json:"ws_connected"`
}

type JsonWorkersObj struct {
    JsonWorkers []string `json:workers`
}

type JsonListeners map[string][]string

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

func (device *FsdbDevice) ExtendSDDL(jsn map[string]interface{}) error {
    // TODO: implement
}

func (device *FsdbDevice) HistoricData(varDef sddl.VarDef, curTime, startTime, endTime time.Time) ([]cloudvar.CloudVarSample, error) {
    // TODO: implement
}

func (device *FsdbDevice) HistoricDataByName(cloudVarName string, curTime, startTime, endTime time.Time) ([]cloudvar.CloudVarSample, error) {
    // TODO: implement
}

func (device *FsdbDevice) HistoricNotifications() ([]Notification, error) {
    // TODO: implement
}

func (device *FsdbDevice) ID() gocql.UUID {
    // TODO: implement
}

func (device *FsdbDevice) IDString() string {
    // TODO: implement
}

func (device *FsdbDevice) InsertSample(varDef sddl.VarDef, t time.Time, value interface{}) error {
    // TODO: implement
}

func (device *FsdbDevice) InsertNotification(notifyType int, t time.Time, msg string) error {
    // TODO: implement
}

func (device *FsdbDevice) LastActivityTime() *time.Time {
    // TODO: implement
}

func (device *FsdbDevice) LatestData(varDef sddl.VarDef) (*cloudvar.CloudVarSample, error) {
    // TODO: implement
}

func (device *FsdbDevice) LatestDataByName(cloudVarName string) (*cloudvar.CloudVarSample, error) {
    // TODO: implement
}

func (device *FsdbDevice) LocationNote() string {
    return device.JsonLocationNote
}

func (device *FsdbDevice) LookupVarDef(cloudVarName string) (sddl.VarDef, error) {
    // TODO: implement
}

func (device *FsdbDevice) Name() string {
    return device.JsonName
}

func (device *FsdbDevice) PublicAccessLevel() AccessLevel {
    // TODO: implement
}

func (device *FsdbDevice) SDDLDocument() sddl.Document {
    return device.doc
}

func (device *FsdbDevice) SDDLDocumentString() string {
    // TODO: implement
}

func (device *FsdbDevice) SecretKey() string {
    return device.JSONSe
}

func (device *FsdbDevice) SetAccountAccess(account Account, access AccessLevel, sharing ShareLevel) error {
    // TODO: implement
}

func (device *FsdbDevice) SetLocationNote(locationNote string) error {
    // TODO: implement
}

func (device *FsdbDevice) SetName(name string) error {
    // TODO: implement
}

func (device *FsdbDevice) SetSDDLDocument(doc sddl.Document) error {
    // TODO: implement
}

func (device *FsdbDevice) UpdateLastActivityTime(t *time.Time) error {
    // TODO: implement
}

func (device *FsdbDevice) UpdateWSConnected(connected bool) error {
    // TODO: implement
}

func (device *FsdbDevice) WSConnected() bool {
    return device.JsonWSConnected
}



func NewDatalayer(cfg config.Config) datalayer.Datalayer {
    return &FsdbDatalayer{cfg: cfg, path: "/var/canopy/db"}
}
