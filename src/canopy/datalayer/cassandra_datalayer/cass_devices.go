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
    "github.com/gocql/gocql"
    "time"
    "canopy/sddl"
    "canopy/canolog"
    "fmt"
)


type CassDevice struct {
    conn *CassConnection
    deviceId gocql.UUID
    name string
    locationNote string
    class *sddl.Class
    classString string
    publicAccessLevel datalayer.AccessLevel
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

func (device *CassDevice) getHistoricData_generic(propname string, datatype sddl.DatatypeEnum, startTime time.Time, endTime time.Time) ([]sddl.PropertySample, error) {
    var timestamp time.Time

    tableName, err := tableNameByDatatype(datatype)
    if err != nil {
        return []sddl.PropertySample{}, err
    }

    query := device.conn.session.Query(`
            SELECT time, value
            FROM ` + tableName + `
            WHERE device_id = ?
                AND propname = ?
    `, device.ID(), propname).Consistency(gocql.One)

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

func (device *CassDevice) HistoricData(property sddl.Property, startTime, endTime time.Time) ([]sddl.PropertySample, error) {
    switch prop := property.(type) {
    case *sddl.Control:
        return device.getHistoricData_generic(prop.Name(), prop.Datatype(), startTime, endTime)
    case *sddl.Sensor:
        return device.getHistoricData_generic(prop.Name(), prop.Datatype(), startTime, endTime)
    default:
        return []sddl.PropertySample{}, fmt.Errorf("HistoricData expects Sensor or Control")
    }
}

func (device *CassDevice) HistoricDataByPropertyName(propertyName string, startTime, endTime time.Time) ([]sddl.PropertySample, error) {
    prop, err := device.LookupProperty(propertyName)
    if err != nil {
        return []sddl.PropertySample{}, err
    }
    return device.HistoricData(prop, startTime, endTime)
}

func (device *CassDevice) ID() gocql.UUID {
    return device.deviceId
}

func (device *CassDevice) HistoricNotifications() ([]datalayer.Notification, error) {
    var uuid gocql.UUID
    var timestamp time.Time
    var dismissed bool
    var msg string
    var notifyType int

    query := device.conn.session.Query(`
            SELECT uuid, time_issued, dismissed, msg, notify_type
            FROM notifications
            WHERE device_id = ?
    `, device.ID()).Consistency(gocql.One)

    iter := query.Iter()
    notifications := []datalayer.Notification{}

    for iter.Scan(&uuid, &timestamp, &dismissed, &msg, &notifyType) {
        notifications = append(notifications, &CassNotification{
                uuid, timestamp, dismissed, msg, notifyType})
    }

    if err := iter.Close(); err != nil {
        return []datalayer.Notification{}, err
    }

    return notifications, nil
}


func (device *CassDevice)InsertNotification(notifyType int, t time.Time, msg string) error {
    err := device.conn.session.Query(`
            INSERT INTO notifications (device_id, time_issued, dismissed, msg, notify_type)
            VALUES (?, ?, false, ?, ?)
    `, device.ID(), t, msg, notifyType).Exec()
    if err != nil {
        return err;
    }
    return nil
}

func (device *CassDevice) insertSensorSample_int(propname string, t time.Time, value int32) error {
    err := device.conn.session.Query(`
            INSERT INTO propval_int (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.ID(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassDevice) insertSensorSample_float(propname string, t time.Time, value float32) error {
    err := device.conn.session.Query(`
            INSERT INTO propval_float (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.ID(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassDevice) insertSensorSample_double(propname string, t time.Time, value float64) error {
    err := device.conn.session.Query(`
            INSERT INTO propval_double (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.ID(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassDevice) insertSensorSample_timestamp(propname string, t time.Time, value time.Time) error {
    err := device.conn.session.Query(`
            INSERT INTO propval_timestamp (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.ID(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassDevice) insertSensorSample_boolean(propname string, t time.Time, value bool) error {
    err := device.conn.session.Query(`
            INSERT INTO propval_boolean (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.ID(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassDevice) insertSensorSample_void(propname string, t time.Time) error {
    err := device.conn.session.Query(`
            INSERT INTO propval_void (device_id, propname, time)
            VALUES (?, ?, ?)
    `, device.ID(), propname, t).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

func (device *CassDevice) insertSensorSample_string(propname string, t time.Time, value string) error {
    err := device.conn.session.Query(`
            INSERT INTO propval_string (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.ID(), propname, t, value).Exec()
    if err != nil {
        return err;
    }
    return nil;
}


func (device *CassDevice) InsertSample(property sddl.Property, t time.Time, value interface{}) error {
    var datatype sddl.DatatypeEnum
    switch prop := property.(type) {
    case *sddl.Control:
        datatype = prop.Datatype()
    case *sddl.Sensor:
        datatype = prop.Datatype()
    default:
        return fmt.Errorf("InsertSample expects control or sensor property")
    }
    propname := property.Name()
    switch datatype {
    case sddl.DATATYPE_VOID:
        return device.insertSensorSample_void(propname, t);
    case sddl.DATATYPE_STRING:
        v, ok := value.(string)
        if !ok {
            return fmt.Errorf("InsertSample expects string value for %s", property.Name())
        }
        return device.insertSensorSample_string(propname, t, v);
    case sddl.DATATYPE_BOOL:
        v, ok := value.(bool)
        if !ok {
            return fmt.Errorf("InsertSample expects bool value for %s", property.Name())
        }
        return device.insertSensorSample_boolean(propname, t, v);
    case sddl.DATATYPE_INT8:
        v, ok := value.(int8)
        if !ok {
            return fmt.Errorf("InsertSample expects int8 value for %s", property.Name())
        }
        return device.insertSensorSample_int(propname, t, int32(v));
    case sddl.DATATYPE_UINT8:
        v, ok := value.(uint8)
        if !ok {
            return fmt.Errorf("InsertSample expects uint8 value for %s", property.Name())
        }
        return device.insertSensorSample_int(propname, t, int32(v));
    case sddl.DATATYPE_INT16:
        v, ok := value.(int16)
        if !ok {
            return fmt.Errorf("InsertSample expects int16 value for %s", property.Name())
        }
        return device.insertSensorSample_int(propname, t, int32(v));
    case sddl.DATATYPE_UINT16:
        v, ok := value.(uint16)
        if !ok {
            return fmt.Errorf("InsertSample expects uint16 value for %s", property.Name())
        }
        return device.insertSensorSample_int(propname, t, int32(v));
    case sddl.DATATYPE_INT32:
        v, ok := value.(int32)
        if !ok {
            return fmt.Errorf("InsertSample expects int32 value for %s", property.Name())
        }
        return device.insertSensorSample_int(propname, t, v);
    case sddl.DATATYPE_UINT32:
        v, ok := value.(uint32)
        if !ok {
            return fmt.Errorf("InsertSample expects uint32 value for %s", property.Name())
        }
        return device.insertSensorSample_int(propname, t, int32(v)) // TODO: verify this works as expected
    case sddl.DATATYPE_FLOAT32:
        v, ok := value.(float32)
        if !ok {
            return fmt.Errorf("InsertSample expects float32 value for %s", property.Name())
        }
        return device.insertSensorSample_float(propname, t, v);
    case sddl.DATATYPE_FLOAT64:
        v, ok := value.(float64)
        if !ok {
            return fmt.Errorf("InsertSample expects float64 value for %s", property.Name())
        }
        return device.insertSensorSample_double(propname, t, v);
    case sddl.DATATYPE_DATETIME:
        v, ok := value.(time.Time)
        if !ok {
            return fmt.Errorf("InsertSample expects time.Time value for %s", property.Name())
        }
        return device.insertSensorSample_timestamp(propname, t, v);
    default:
        return fmt.Errorf("InsertSample unsupported datatype ", datatype)
    }
}

func (device *CassDevice) getLatestData_generic(propname string, datatype sddl.DatatypeEnum) (*sddl.PropertySample, error) {
    var timestamp time.Time
    var sample *sddl.PropertySample

    tableName, err := tableNameByDatatype(datatype)
    if err != nil {
        return nil, err
    }

    query := device.conn.session.Query(`
            SELECT time, value
            FROM ` + tableName + `
            WHERE device_id = ?
                AND propname = ?
            ORDER BY time DESC
            LIMIT 1`, device.ID(), propname).Consistency(gocql.One)

    switch datatype {
    case sddl.DATATYPE_VOID:
        err = query.Scan(&timestamp)
        sample = &sddl.PropertySample{timestamp, nil}
    case sddl.DATATYPE_STRING:
        var value string
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_BOOL:
        var value bool
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_INT8:
        var value int8
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_UINT8:
        var value uint8
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_INT16:
        var value int16
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_UINT16:
        var value uint16
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_INT32:
        var value int32
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_UINT32:
        var value uint32
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_FLOAT32:
        var value float32
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_FLOAT64:
        var value float64
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_DATETIME:
        var value time.Time
        err = query.Scan(&timestamp, &value)
        sample = &sddl.PropertySample{timestamp, value}
    case sddl.DATATYPE_INVALID:
        return nil, fmt.Errorf("Cannot get property values for DATATYPE_INVALID");
    default:
        return nil, fmt.Errorf("Cannot get property values for datatype %d", datatype);
    }

    if err != nil {
        return nil, fmt.Errorf("Error reading latest property value: ", err)
    }

    return sample, nil
}

func (device *CassDevice) LatestData(property sddl.Property) (*sddl.PropertySample, error) {
    switch prop := property.(type) {
    case *sddl.Control:
        return device.getLatestData_generic(prop.Name(), prop.Datatype())
    case *sddl.Sensor:
        return device.getLatestData_generic(prop.Name(), prop.Datatype())
    default:
        return nil, fmt.Errorf("LatestData expects Sensor or Control")
    }
}


func (device *CassDevice) LatestDataByPropertyName(propertyName string) (*sddl.PropertySample, error) {
    prop, err := device.LookupProperty(propertyName)
    if err != nil {
        return nil, err
    }
    return device.LatestData(prop)
}

func (device *CassDevice) LookupProperty(propName string) (sddl.Property, error) {
    sddlClass := device.SDDLClass()
    if sddlClass == nil {
        return nil, fmt.Errorf("Cannot lookup property %s, device %s has unknown SDDL", propName, device.Name())
    }

    return sddlClass.LookupProperty(propName)
}

func (device *CassDevice) Name() string {
    return device.name
}

func (device *CassDevice) SDDLClass() *sddl.Class {
    return device.class
}

func (device *CassDevice) SDDLClassString() string{
    return device.classString
}

func (device *CassDevice) SetAccountAccess(account datalayer.Account, access datalayer.AccessLevel, sharing datalayer.ShareLevel) error {
    /* TODO: Incorporate sharing level */
    err := device.conn.session.Query(`
            INSERT INTO device_permissions (username, device_id, access_level)
            VALUES (?, ?, ?)
    `, account.Username(), device.ID(), access).Exec()

    return err
}

func (device *CassDevice) SetLocationNote(locationNote string) error {
    err := device.conn.session.Query(`
            UPDATE devices
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
            UPDATE devices
            SET friendly_name = ?
            WHERE device_id = ?
    `, name, device.ID()).Exec()
    if err != nil {
        return err;
    }
    device.name = name;
    return nil;
}


func (device *CassDevice) SetSDDLClass(class *sddl.Class) error {
    sddlText, err := class.ToString()
    if err != nil {
        return err
    }
    canolog.Info("sddlText: ", sddlText)

    err = device.conn.session.Query(`
            UPDATE devices
            SET sddl = ?
            WHERE device_id = ?
    `, sddlText, device.ID()).Exec()
    if err != nil {
        return err;
    }
    return nil;
}

