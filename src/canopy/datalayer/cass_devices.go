package datalayer

import (
    "github.com/gocql/gocql"
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

/*func (dl *CassandraDatalayer) (GetSensorData(device_uuid gocql.UUID, propname string, startTime time.Time, endTime time.Time) {
    if err := dl.session.Query(`
            INSERT INTO devices (device_id, friendly_name)
            VALUES (?, ?)
    `, gocql.TimeUUID(), friendlyName).Exec(); err != nil {
        log.Print(err)
    }
}*/
