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
package datalayer

import (
    "github.com/gocql/gocql"
    "time"
    "canopy/sddl"
    "fmt"
)

type AccessLevel int
const (
    NoAccess = iota
    ReadOnlyAccess
    ReadWriteAccess
    ReadWriteShareAccess
)

type CassandraDevice struct {
    dl *CassandraDatalayer
    deviceId gocql.UUID
    friendlyName string
    class *sddl.Class
    classString string
}

type SensorSample struct {
    Timestamp time.Time
    Value float64
}

func (dl *CassandraDatalayer) CreateDevice(friendlyName string) (*CassandraDevice, error) {
    deviceId := gocql.TimeUUID()

    if err := dl.session.Query(`
            INSERT INTO devices (device_id, friendly_name)
            VALUES (?, ?)
    `, deviceId, friendlyName).Exec(); err != nil {
        return nil, err
    }
    return &CassandraDevice{dl, deviceId, friendlyName, nil, ""}, nil
}

func (dl *CassandraDatalayer) LookupDeviceByStringId(id string) (*CassandraDevice, error) {
    deviceId, err := gocql.ParseUUID(id)
    if err != nil {
        return nil, err
    }
    return dl.LookupDevice(deviceId)
}

func (dl *CassandraDatalayer) LookupDevice(deviceId gocql.UUID) (*CassandraDevice, error) {
    var device CassandraDevice

    device.deviceId = deviceId
    device.dl = dl

    err := dl.session.Query(`
        SELECT friendly_name, sddl
        FROM devices
        WHERE device_id = ?
        LIMIT 1`, deviceId).Consistency(gocql.One).Scan(
            &device.friendlyName,
            &device.classString)
    if err != nil {
        return nil, err
    }

    if device.classString != "" {
        device.class, err = sddl.ParseClassString("anonymous", device.classString)
        if err != nil {
            return nil, err
        }
    }

    return &device, nil
}

func (device *CassandraDevice) GetId() gocql.UUID {
    return device.deviceId
}

func (device *CassandraDevice) GetFriendlyName() string {
    return device.friendlyName
}

func (device *CassandraDevice) SetAccountAccess(account *CassandraAccount, level AccessLevel) error {
    err := device.dl.session.Query(`
            INSERT INTO device_permissions (username, device_id, access_level)
            VALUES (?, ?, ?)
    `, account.Username(), device.GetId(), level).Exec()

    return err
}

func (device *CassandraDevice) SetFriendlyName(friendlyName string) error {
    err := device.dl.session.Query(`
            UPDATE devices
            SET friendly_name = ?
            WHERE device_id = ?
    `, friendlyName, device.GetId()).Exec()
    if err != nil {
        return err;
    }
    device.friendlyName = friendlyName;
    return nil;
}

func (device *CassandraDevice) SetLocationNote(locationNote string) error {
    /*err = device.dl.session.Query(`
            UPDATE devices
            SET location_note = ?
            WHERE device_id = ?
    `, locationNote, device.GetId()).Exec()
    if err != nil {
        return err;
    }
    device.locationNote = locationNote;*/
    return nil;
}

func (device *CassandraDevice) insertSensorSample_int(propname string, t time.Time, value int32) error {
    err := device.dl.session.Query(`
            INSERT INTO propval_int (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.GetId(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) insertSensorSample_float(propname string, t time.Time, value float32) error {
    err := device.dl.session.Query(`
            INSERT INTO propval_float (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.GetId(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) insertSensorSample_double(propname string, t time.Time, value float64) error {
    err := device.dl.session.Query(`
            INSERT INTO propval_double (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.GetId(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) insertSensorSample_timestamp(propname string, t time.Time, value time.Time) error {
    err := device.dl.session.Query(`
            INSERT INTO propval_timestamp (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.GetId(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) insertSensorSample_boolean(propname string, t time.Time, value bool) error {
    err := device.dl.session.Query(`
            INSERT INTO propval_boolean (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.GetId(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) insertSensorSample_void(propname string, t time.Time) error {
    err := device.dl.session.Query(`
            INSERT INTO propval_void (device_id, propname, time)
            VALUES (?, ?, ?)
    `, device.GetId(), propname, t).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) insertSensorSample_string(propname string, t time.Time, value string) error {
    err := device.dl.session.Query(`
            INSERT INTO propval_string (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.GetId(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) InsertSensorSample_void(propname string, t time.Time) error {
    return device.insertSensorSample_void(propname, t);
}
func (device *CassandraDevice) InsertSensorSample_string(propname string, t time.Time, value string) error {
    return device.insertSensorSample_string(propname, t, value);
}
func (device *CassandraDevice) InsertSensorSample_bool(propname string, t time.Time, value bool) error {
    return device.insertSensorSample_boolean(propname, t, value);
}
func (device *CassandraDevice) InsertSensorSample_int8(propname string, t time.Time, value int8) error {
    return device.insertSensorSample_int(propname, t, int32(value));
}
func (device *CassandraDevice) InsertSensorSample_uint8(propname string, t time.Time, value uint8) error {
    return device.insertSensorSample_int(propname, t, int32(value));
}
func (device *CassandraDevice) InsertSensorSample_int16(propname string, t time.Time, value int16) error {
    return device.insertSensorSample_int(propname, t, int32(value));
}
func (device *CassandraDevice) InsertSensorSample_uint16(propname string, t time.Time, value uint16) error {
    return device.insertSensorSample_int(propname, t, int32(value));
}
func (device *CassandraDevice) InsertSensorSample_int32(propname string, t time.Time, value int32) error {
    return device.insertSensorSample_int(propname, t, int32(value));
}
func (device *CassandraDevice) InsertSensorSample_uint32(propname string, t time.Time, value uint32) error {
    return device.insertSensorSample_int(propname, t, int32(value)); /* TODO: verify this works  as expected */
}
func (device *CassandraDevice) InsertSensorSample_float32(propname string, t time.Time, value float32) error {
    return device.insertSensorSample_float(propname, t, value);
}
func (device *CassandraDevice) InsertSensorSample_float64(propname string, t time.Time, value float64) error {
    return device.insertSensorSample_double(propname, t, value);
}
func (device *CassandraDevice) InsertSensorSample_datetime(propname string, t time.Time, value time.Time) error {
    return device.insertSensorSample_timestamp(propname, t, value);
}

func (device *CassandraDevice) SetSDDLClass(class *sddl.Class) error {
    sddlText, err := class.ToString()
    if err != nil {
        return err
    }

    err = device.dl.session.Query(`
            UPDATE devices
            SET sddl = ?
            WHERE device_id = ?
    `, sddlText, device.GetId()).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassandraDevice) GetSensorData(propname string, startTime time.Time, endTime time.Time) ([]SensorSample, error) {
    var value float64
    var timestamp time.Time
    /* TODO: restrict between startTime and endTime */
    query := device.dl.session.Query(`
            SELECT time, value
            FROM sensor_data
            WHERE device_id = ?
                AND propname = ?
    `, device.GetId(), propname).Consistency(gocql.One);

    iter := query.Iter()
    samples := []SensorSample{}
    for iter.Scan(&timestamp, &value) {
        samples = append(samples, SensorSample{timestamp, value})
    }
    if err := iter.Close(); err != nil {
        return []SensorSample{}, err
    }

    return samples, nil
}

func tableNameByDatatype(datatype sddl.DatatypeEnum) (string, error) {
    switch datatype {
    case sddl.DATATYPE_VOID:
        return "propval_void", nil
    case sddl.DATATYPE_STRING:
        return "propval_string", nil
    case sddl.DATATYPE_BOOL:
        return "propval_boolean", nil
    case sddl.DATATYPE_INT8:
        return "propval_int", nil
    case sddl.DATATYPE_UINT8:
        return "propval_int", nil
    case sddl.DATATYPE_INT16:
        return "propval_int", nil
    case sddl.DATATYPE_UINT16:
        return "propval_int", nil
    case sddl.DATATYPE_INT32:
        return "propval_int", nil
    case sddl.DATATYPE_UINT32:
        return "propval_int", nil
    case sddl.DATATYPE_FLOAT32:
        return "propval_float", nil
    case sddl.DATATYPE_FLOAT64:
        return "propval_double", nil
    case sddl.DATATYPE_DATETIME:
        return "propval_timestamp", nil
    case sddl.DATATYPE_INVALID:
        return "", fmt.Errorf("DATATYPE_INVALID not allowed in tableNameByDatatype");
    default: 
        return "", fmt.Errorf("Unexpected datatype in tableNameByDatatype: %d", datatype);
    }
}

func (device *CassandraDevice) getSensorData_generic(propname string, datatype sddl.DatatypeEnum, startTime time.Time, endTime time.Time) ([]sddl.PropertySample, error) {
    var timestamp time.Time

    tableName, err := tableNameByDatatype(datatype)
    if err != nil {
        return []sddl.PropertySample{}, err
    }

    query := device.dl.session.Query(`
            SELECT time, value
            FROM ` + tableName + `
            WHERE device_id = ?
                AND propname = ?
    `, device.GetId(), propname).Consistency(gocql.One)

    iter := query.Iter()
    samples := []sddl.PropertySample{}

    switch datatype {
    case sddl.DATATYPE_VOID:
        var value interface{}
        for iter.Scan(&timestamp) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_STRING:
        var value string
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_BOOL:
        var value bool
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_INT8:
        var value int8
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_UINT8:
        var value uint8
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_INT16:
        var value int16
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_UINT16:
        var value uint16
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_INT32:
        var value int32
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_UINT32:
        var value uint32
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_FLOAT32:
        var value float32
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_FLOAT64:
        var value float64
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_DATETIME:
        var value time.Time
        for iter.Scan(&timestamp, &value) {
            samples = append(samples, sddl.PropertySample{timestamp, value})
        }
    case sddl.DATATYPE_INVALID:
        return []sddl.PropertySample{}, fmt.Errorf("Cannot get property values for DATATYPE_INVALID");
    default:
        return []sddl.PropertySample{}, fmt.Errorf("Cannot get property values for datatype %d", datatype);
    }

    if err := iter.Close(); err != nil {
        return []sddl.PropertySample{}, err
    }

    return samples, nil
}

func (device *CassandraDevice) GetPropertyData(property sddl.Property, startTime time.Time, endTime time.Time) ([]sddl.PropertySample, error) {
    switch prop := property.(type) {
    case *sddl.Control:
        return device.getSensorData_generic(prop.Name(), prop.Datatype(), startTime, endTime)
    case *sddl.Sensor:
        return device.getSensorData_generic(prop.Name(), prop.Datatype(), startTime, endTime)
    default:
        return []sddl.PropertySample{}, fmt.Errorf("GetPropertyData expects Sensor or Control")
    }
}


func (device *CassandraDevice) GetPropertyDataByName(propName string, startTime time.Time, endTime time.Time) ([]sddl.PropertySample, error) {
    prop, err := device.LookupProperty(propName)
    if err != nil {
        return []sddl.PropertySample{}, err
    }
    return device.GetPropertyData(prop, startTime, endTime)
}

func (device *CassandraDevice) LookupProperty(propName string) (sddl.Property, error) {
    sddlClass := device.SDDLClass()
    if sddlClass == nil {
        return nil, fmt.Errorf("Cannot lookup property %s, device %s has unknown SDDL", propName, device.GetFriendlyName())
    }

    return sddlClass.LookupProperty(propName)
}

func (device *CassandraDevice) GetCurrentSensorData(propname string) (*SensorSample, error) {
    var value float64
    var timestamp time.Time

    err := device.dl.session.Query(`
        SELECT time, value
        FROM sensor_data
        WHERE device_id = ?
            AND propname = ?
        ORDER BY propname DESC
        LIMIT 1`, device.GetId(), propname).Consistency(gocql.One).Scan(
            &timestamp, 
            &value)
    if err != nil {
        return nil, err
    }

    sample := SensorSample{timestamp, value}
    return &sample, nil
}

func (device *CassandraDevice) SDDLClass() *sddl.Class {
    return device.class
}

func (device *CassandraDevice) SDDLClassString() string{
    return device.classString
}
