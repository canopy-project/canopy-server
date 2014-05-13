package datalayer

import (
    "github.com/gocql/gocql"
    "log"
    //"time"
)

func (dl *CassandraDatalayer) CreateDevice(friendlyName string) {
    if err := dl.session.Query(`
            INSERT INTO devices (device_id, friendly_name)
            VALUES (?, ?)
    `, gocql.TimeUUID(), friendlyName).Exec(); err != nil {
        log.Print(err)
    }
}

/*func (dl *CassandraDatalayer) (GetSensorData(device_uuid gocql.UUID, propname string, startTime time.Time, endTime time.Time) {
    if err := dl.session.Query(`
            INSERT INTO devices (device_id, friendly_name)
            VALUES (?, ?)
    `, gocql.TimeUUID(), friendlyName).Exec(); err != nil {
        log.Print(err)
    }
}*/
