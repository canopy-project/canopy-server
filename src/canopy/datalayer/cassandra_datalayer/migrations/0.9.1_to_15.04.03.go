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
    "canopy/canolog"
    "github.com/gocql/gocql"
)

var migrationQueries_0_9_1_to_15_04_03 []string = []string{
    `CREATE TABLE var_lastupdatetime (
        device_id uuid,
        var_name text,
        last_update timestamp,
        PRIMARY KEY(device_id, var_name)
    ) WITH COMPACT STORAGE`,

    // Keeps track of which buckets have been created for use by garbage
    // collector.
    `CREATE TABLE var_buckets (
        device_id uuid,
        var_name text,
        lod int,
        timeprefix text,
        endtime timestamp,
        PRIMARY KEY((device_id, var_name, lod), timeprefix)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  uint8
    //  int8
    //  int16
    //  uint16
    //  int32
    //  uint32
    `CREATE TABLE varsample_int (
        device_id uuid,
        propname text,
        timeprefix text,
        time timestamp,
        value int,
        PRIMARY KEY((device_id, propname, timeprefix), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  float32
    `CREATE TABLE varsample_float (
        device_id uuid,
        propname text,
        timeprefix text,
        time timestamp,
        value float,
        PRIMARY KEY((device_id, propname, timeprefix), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  float64
    `CREATE TABLE varsample_double (
        device_id uuid,
        propname text,
        timeprefix text,
        time timestamp,
        value double,
        PRIMARY KEY((device_id, propname, timeprefix), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  datetime
    `CREATE TABLE varsample_timestamp (
        device_id uuid,
        propname text,
        timeprefix text,
        time timestamp,
        value timestamp,
        PRIMARY KEY((device_id, propname, timeprefix), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  bool
    `CREATE TABLE varsample_boolean (
        device_id uuid,
        propname text,
        timeprefix text,
        time timestamp,
        value timestamp,
        PRIMARY KEY((device_id, propname, timeprefix), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  string
    `CREATE TABLE varsample_string (
        device_id uuid,
        propname text,
        timeprefix text,
        time timestamp,
        value text,
        PRIMARY KEY((device_id, propname, timeprefix), time)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE workers (
        name text,
        status text,
        PRIMARY KEY(name)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE listeners (
        key text,
        workers set<text>,
        PRIMARY KEY(key)
    ) `,

    `ALTER TABLE devices ADD location_note text`,
    `ALTER TABLE devices ADD ws_connected boolean`,
}

func Migrate_0_9_1_to_15_04_03(session *gocql.Session) error {
    // Perform all migration queries.
    for _, query := range migrationQueries_0_9_1_to_15_04_03 {
        canolog.Info(query)
        if err := session.Query(query).Exec(); err != nil {
            // Ignore errors (just print them).
            canolog.Warn(query, ": ", err)
        }
    }
    return nil
}
