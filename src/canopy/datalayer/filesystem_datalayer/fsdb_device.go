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


func (device *FsdbDevice) ExtendSDDL(jsn map[string]interface{}) error {
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
    return gocql.ParseUUID(device)
}

func (device *FsdbDevice) IDString() string {
    return device.JsonID
}

func (device *FsdbDevice) InsertSample(varDef sddl.VarDef, t time.Time, value interface{}) error {
    // TODO: implement
}

func (device *FsdbDevice) InsertNotification(notifyType int, t time.Time, msg string) error {
    return fmt.Errorf("Not Implemented")
}

func (device *FsdbDevice) LastActivityTime() *time.Time {
    return device.JsonLastSeen
}

func (device *FsdbDevice) LatestData(varDef sddl.VarDef) (*cloudvar.CloudVarSample, error) {
    // TODO: implement
}

func (device *FsdbDevice) LatestDataByName(cloudVarName string) (*cloudvar.CloudVarSample, error) {
    varDef, err := device.LookupVarDef(varName)
    if err != nil {
        return nil, err
    }
    return device.LatestData(varDef)
}

func (device *FsdbDevice) LocationNote() string {
    return device.JsonLocationNote
}

func (device *FsdbDevice) LookupVarDef(cloudVarName string) (sddl.VarDef, error) {
    doc := device.SDDLDocument()

    if doc == nil {
        return nil, fmt.Errorf("Cannot lookup property %s, device %s has unknown SDDL", varName, device.Name())
    }

    return doc.LookupVarDef(varName)
}

func (device *FsdbDevice) Name() string {
    return device.JsonName
}

func (device *FsdbDevice) PublicAccessLevel() AccessLevel {
    return device.JsonPublicAccessLevel
}

func (device *FsdbDevice) SDDLDocument() sddl.Document {
    return device.doc
}

func (device *FsdbDevice) SDDLDocumentString() string {
    out, _ := device.doc.ToString()
    return out
}

func (device *FsdbDevice) SecretKey() string {
    return device.JsonSecretKey
}

func (device *FsdbDevice) SetAccountAccess(account Account, access AccessLevel, sharing ShareLevel) error {
    // TODO: implement
}

func (device *FsdbDevice) SetLocationNote(locationNote string) error {
    // TODO: Input validation
    // Modify
    device.JsonLocationNote = location_note

    // Write
    return device.save()
}

func (device *FsdbDevice) SetName(name string) error {
    // TODO: Input validation
    // Modify
    device.JsonName = name

    // Write
    return device.save()
}

func (device *FsdbDevice) SetSDDLDocument(doc sddl.Document) error {
    // Modify
    device.doc = doc

    // Write
    device.save()
}

func (device *FsdbDevice) UpdateLastActivityTime(t *time.Time) error {
    // Modify
    device.JsonLastSeen = t

    // Write
    return device.save()
}

func (device *FsdbDevice) UpdateWSConnected(connected bool) error {
    // Modify
    device.JsonWSConnected = connected

    // Write
    return device.save()
}

func (device *FsdbDevice) WSConnected() bool {
    return device.JsonWSConnected
}

