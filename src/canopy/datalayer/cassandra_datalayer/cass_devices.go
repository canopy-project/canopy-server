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

import (
    "canopy/datalayer"
    "github.com/gocql/gocql"
    "time"
    "canopy/sddl"
    "canopy/canolog"
    "canopy/cloudvar"
    "fmt"
)


type CassDevice struct {
    conn *CassConnection
    deviceId string
    doc sddl.Document
    docString string
    last_seen *time.Time
    locationNote string
    name string
    publicAccessLevel datalayer.AccessLevel
    secretKey string
    wsConnected bool
}

func (device *CassDevice) ExtendSDDL(jsn map[string]interface{}) error {
    // TODO: Race condition?
    doc := device.SDDLDocument()

    err := doc.Extend(jsn)
    if err != nil {
        canolog.Error("Error extending class ", jsn, err)
        return err
    }

    // save modified SDDL class to DB
    err = device.SetSDDLDocument(doc)
    if err != nil {
        canolog.Error("Error saving SDDL: ", err)
        return err
    }
    return nil
}

/*func (device *CassDevice) HistoricData(varDef sddl.VarDef, startTime, endTime time.Time) ([]cloudvar.CloudVarSample, error) {
    return device.getHistoricData_generic(varDef.Name(), varDef.Datatype(), startTime, endTime)
}*/

func (device *CassDevice) HistoricDataByName(cloudVarName string, curTime, startTime, endTime time.Time) ([]cloudvar.CloudVarSample, error) {
    varDef, err := device.LookupVarDef(cloudVarName)
    if err != nil {
        return []cloudvar.CloudVarSample{}, err
    }
    return device.HistoricData(varDef, curTime, startTime, endTime)
}

func (device *CassDevice) ID() string{
    return device.deviceId
}

func (device *CassDevice) SecretKey() string {
    return device.secretKey
}

func (device *CassDevice) HistoricNotifications() ([]datalayer.Notification, error) {
    var deviceId string
    var timestamp time.Time
    var dismissed bool
    var msg string
    var notifyType int

    query := device.conn.session.Query(`
            SELECT device_id, time_issued, dismissed, msg, notify_type
            FROM notifications_v2
            WHERE device_id = ?
    `, device.ID()).Consistency(gocql.One)

    iter := query.Iter()
    notifications := []datalayer.Notification{}

    for iter.Scan(&deviceId, &timestamp, &dismissed, &msg, &notifyType) {
        notifications = append(notifications, &CassNotification{
                deviceId, timestamp, dismissed, msg, notifyType})
    }

    if err := iter.Close(); err != nil {
        return []datalayer.Notification{}, err
    }

    return notifications, nil
}


func (device *CassDevice)InsertNotification(notifyType int, t time.Time, msg string) error {
    err := device.conn.session.Query(`
            INSERT INTO notifications_v2 (device_id, time_issued, dismissed, msg, notify_type)
            VALUES (?, ?, false, ?, ?)
    `, device.ID(), t, msg, notifyType).Exec()
    if err != nil {
        return err;
    }
    return nil
}

func (device *CassDevice) LastActivityTime() *time.Time {
    return device.last_seen
}

func (device *CassDevice) LatestDataByName(varName string) (*cloudvar.CloudVarSample, error) {
    varDef, err := device.LookupVarDef(varName)
    if err != nil {
        return nil, err
    }
    return device.LatestData(varDef)
}

func (device *CassDevice) LocationNote() string {
    return device.locationNote
}

func (device *CassDevice) LookupVarDef(varName string) (sddl.VarDef, error) {
    doc := device.SDDLDocument()

    if doc == nil {
        return nil, fmt.Errorf("Cannot lookup property %s, device %s has unknown SDDL", varName, device.Name())
    }

    return doc.LookupVarDef(varName)
}

func (device *CassDevice) Name() string {
    return device.name
}

func (device *CassDevice) PublicAccessLevel() datalayer.AccessLevel {
    return device.publicAccessLevel
}

func (device *CassDevice) SDDLDocument() sddl.Document {
    return device.doc
}

func (device *CassDevice) SDDLDocumentString() string {
    return device.docString
}

func (device *CassDevice) SetAccountAccess(account datalayer.Account, access datalayer.AccessLevel, sharing datalayer.ShareLevel) error {
    /* TODO: Incorporate sharing level */
    err := device.conn.session.Query(`
            INSERT INTO device_permissions (username, device_id, access_level)
            VALUES (?, ?, ?)
    `, account.Username(), device.ID(), access).Exec()

    return err
}

func (device *CassDevice) SetTeamAccess(org datalayer.Organization, team string, access datalayer.AccessLevel, sharing datalayer.ShareLevel) error {
    org2, ok := org.(*CassOrganization)
    if !ok {
        return fmt.Errorf("Type error in SetTeamAccess")
    }
    /* TODO: Incorporate sharing level */
    err := device.conn.session.Query(`
            INSERT INTO device_team_permissions (org_id, team, device_id, access_level)
            VALUES (?, ?, ?, ?)
    `, org2.id, team, device.ID(), access).Exec()

    return err
}

func (device *CassDevice) SetLocationNote(locationNote string) error {
    err := device.conn.session.Query(`
            UPDATE devices_v2
            SET location_note = ?
            WHERE device_id = ?
    `, locationNote, device.ID()).Exec()
    if err != nil {
        return err;
    }
    device.locationNote = locationNote
    return nil;
}

func (device *CassDevice) SetName(name string) error {
    err := device.conn.session.Query(`
            UPDATE devices_v2
            SET friendly_name = ?
            WHERE device_id = ?
    `, name, device.ID()).Exec()
    if err != nil {
        return err;
    }
    device.name = name;
    return nil;
}


func (device *CassDevice) SetSDDLDocument(doc sddl.Document) error {
    sddlText, err := doc.ToString()
    if err != nil {
        return err
    }

    err = device.conn.session.Query(`
            UPDATE devices_v2
            SET sddl = ?
            WHERE device_id = ?
    `, sddlText, device.ID()).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassDevice) UpdateLastActivityTime(tp *time.Time) error {
    var t time.Time
    if tp == nil {
        t = time.Now()
    } else {
        t = *tp
    }
    err := device.conn.session.Query(`
            UPDATE devices_v2
            SET last_seen = ?
            WHERE device_id = ?
    `, t, device.ID()).Exec()
    if err != nil {
        return err;
    }
    device.last_seen = &t
    return nil;
}

func (device *CassDevice) UpdateWSConnected(connected bool) error {
    err := device.conn.session.Query(`
            UPDATE devices_v2
            SET ws_connected = ?
            WHERE device_id = ?
    `, connected, device.ID()).Exec()
    if err != nil {
        return err;
    }
    device.wsConnected = connected
    return nil;
}

func (device *CassDevice) WSConnected() bool {
    return device.wsConnected
}
