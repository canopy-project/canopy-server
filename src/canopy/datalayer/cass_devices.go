package datalayer

import (
    "github.com/gocql/gocql"
    "time"
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
    return &CassandraDevice{dl, deviceId, friendlyName}, nil
}

func (dl *CassandraDatalayer) LookupDevice(deviceId gocql.UUID) (*CassandraDevice, error) {
    var device CassandraDevice

    device.deviceId = deviceId
    device.dl = dl

    err := dl.session.Query(`
        SELECT friendly_name
        FROM devices
        WHERE device_id = ?
        LIMIT 1`, deviceId).Consistency(gocql.One).Scan(&device.friendlyName)
    if err != nil {
        return nil, err
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
    `, account.GetUsername(), device.GetId(), level).Exec()

    return err
}

func (device *CassandraDevice) InsertSensorSample(propname string, t time.Time, value float64) error {
    err := device.dl.session.Query(`
            INSERT INTO sensor_data (device_id, propname, time, value)
            VALUES (?, ?, ?, ?)
    `, device.GetId(), propname, t, value).Exec()
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

