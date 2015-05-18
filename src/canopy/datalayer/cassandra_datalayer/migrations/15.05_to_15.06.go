// Copright 2014-2015 Canopy Services, Inc.
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
package migrations

import (
    "fmt"
    "github.com/gocql/gocql"
    "time"
)

func migrate_lastupdatetime(session *gocql.Session) {
    var device_uuid gocql.UUID
    var var_name string
    var last_update time.Time

    query := session.Query(`
        SELECT device_id, var_name, last_update
        FROM var_lastupdatetime`).Consistency(gocql.One)
    iter := query.Iter()
    for iter.Scan(&device_uuid, &var_name, &last_update) {
        err := session.Query(`
            INSERT INTO var_lastupdatetime_v2
                (device_id, var_name, last_update)
            VALUES (?, ?, ?)
        `, device_uuid.String(), var_name, last_update).Exec()
        if err != nil {
            fmt.Println("Err:", err)
            fmt.Println("ABORT")
            return
        }
    }
    err := iter.Close();
    if err != nil {
        fmt.Println("Err: ", err);
    }
    fmt.Println("DONE migrating var_lastupdatetime")
}

func migrate_devices(session *gocql.Session) {
    var device_uuid gocql.UUID
    var secret_key string
    var friendly_name string
    var sddl string
    var public_access_level int
    var last_seen time.Time
    var location_note string
    var ws_connected bool

    query := session.Query(`
        SELECT device_id, secret_key, friendly_name, sddl, public_access_level, last_seen, location_note, ws_connected
        FROM devices`).Consistency(gocql.One)
    iter := query.Iter()
    for iter.Scan(&device_uuid, &secret_key, &friendly_name, &sddl, &public_access_level, &last_seen, &location_note, &ws_connected) {
        err := session.Query(`
            INSERT INTO devices_v2
                (device_id, secret_key, friendly_name, sddl, public_access_level, last_seen, location_note, ws_connected)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `, device_uuid.String(), secret_key, friendly_name, sddl, public_access_level, last_seen, location_note, ws_connected).Exec()
        if err != nil {
            fmt.Println("Err:", err)
            fmt.Println("ABORT")
            return
        }
    }
    err := iter.Close();
    if err != nil {
        fmt.Println("Err: ", err);
    }
    fmt.Println("DONE migrating devices")
}

func migrate_device_permissions(session *gocql.Session) {
    var username string
    var device_uuid gocql.UUID
    var access_level int

    query := session.Query(`
        SELECT username, device_id, access_level
        FROM device_permissions`).Consistency(gocql.One)
    iter := query.Iter()
    for iter.Scan(&username, &device_uuid, &access_level) {
        err := session.Query(`
            INSERT INTO device_permissions_v2
                (username, device_id, access_level)
            VALUES (?, ?, ?)
        `, username, device_uuid.String(), access_level).Exec()
        if err != nil {
            fmt.Println("Err:", err)
            fmt.Println("ABORT")
            return
        }
    }
    err := iter.Close();
    if err != nil {
        fmt.Println("Err: ", err);
    }
    fmt.Println("DONE migrating device_permissions")
}

func Migrate_15_05_to_15_06(session *gocql.Session) error {
    // Copy data from old tables to new tables

    // DONE:
    // migrate_lastupdatetime(session)
    //migrate_devices(session)
    //migrate_device_permissions(session)

    // NOT DONE:
    // var_buckets_v2
    // varsample_int_v2
    // varsample_float_v2
    // varsample_double_v2
    // varsample_timestamp_v2
    // varsample_boolean_v2
    // varsample_void_v2
    // varsample_string_v2
    // notifications_v2
    return nil
}
