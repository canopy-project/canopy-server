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
    "canopy/canolog"
    "canopy/config"
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer/migrations"
    "fmt"
    "github.com/gocql/gocql"
)


//
// Cassandra stores data in column families (aka tables).  Each column family
// (table) has multiple rows.  Each row has a row key.  Each row also has an
// internal table of key-value pairs (aka "internal rows" or "cells").
// 
//  COLUMN FAMILY
//
//      +--------------------------------+
//      | [ROW_KEY0]                     |
//      |      KEY0: VALUE0              |
//      |      KEY1: VALUE1              |
//      |      ...                       |
//      +--------------------------------+
//      | [ROW_KEY1]                     |
//      |      KEY0: VALUE0              |
//      |      KEY1: VALUE1              |
//      |      ...                       |
//      +--------------------------------+
//      |                                |
//
// The internal keys are stored in sorted order within a row.  A row's contents
// is never split across nodes.
//
// For storing simple data, we could use:
//
//      CREATE TABLE propval_<datatype> (
//          device_id uuid,
//          propname text,
//          time timestamp,
//          value <datatype>,
//          PRIMARY KEY (device_id, propname, time)
//      ) WITH COMPACT STORAGE
//
//  Which maps to (for example):
//
//      propval_int
//
//      +---------------------------------+
//      | device_id (row key)             |
//      |      propname|timestamp : value |
//      |      propname|timestamp : value |
//      |      propname|timestamp : value |
//      +---------------------------------+
//      |                                 |
//
// Instead we should use:
//
//      CREATE TABLE propval_<datatype> (
//          device_id uuid,
//          propname text,
//          time timestamp,
//          value <datatype>,
//          PRIMARY KEY ((device_id, propname), time)
//      ) WITH COMPACT STORAGE
//
//  Which maps to (for example):
//
//      propval_int
//
//      +---------------------------------+
//      | device_id|propname (row key)    |
//      |      timestamp : value          |
//      |      timestamp : value          |
//      |      timestamp : value          |
//      +---------------------------------+
//      |                                 |
//
//      Note that the concatenation of property name and timestamp is used as
//      the internal keys.
//
//
//  In theory we could put this all in a single column family (rather than
//  having a separate one for each datatype).  However, CQL does not appear to
//  have the flexibility to do this efficiently.  If we tried, for example:
//
//      CREATE TABLE propval_<datatype> (
//          device_id uuid,
//          propname text,
//          time timestamp,
//          value_int int,
//          value_bigint bigint,
//          value_string text,
//          PRIMARY KEY (device_id, propname, time)
//      ) WITH COMPACT STORAGE
//
//
//  The result would be:
//
//      +------------------------------------------------+
//      | device_id (row key)                            |
//      |      propname|timestamp|"value_int" : value    |
//      |      propname|timestamp|"value_int" : value    |
//      |      propname|timestamp|"value_int" : value    |
//      |                                                |
//      |      propname|timestamp|"value_string" : value |
//      |      propname|timestamp|"value_string" : value |
//      +------------------------------------------------+
//      |                                                |
//
//  Which is not nearly as efficient, because it would store, literaly, the
//  word "value_int" alongside each 32-bit integer data sample.
//
//  So instead, we create a separate table for each datatype.
//
//  DATA ROLLUP (LOD & DATA TRIMMING)
//
//  Another consideration is that we want to be able to:
//      - Delete old data.
//      - Quickly lookup "low res" data over long time ranges (i.e.: give me 1
//      sample/day for each day last year).
//
//  To achieve this, we put part of the timestamp in the row key:
//
//      CREATE TABLE propval_<datatype> (
//          device_id uuid,
//          propname text,
//          timeprefix text,
//          time timestamp,
//          value <datatype>,
//          PRIMARY KEY ((device_id, propname, timeprefix), time)
//      ) WITH COMPACT STORAGE
//
//      +--------------------------------------------+
//      | device_id|propname|timeprefix (row key)    |
//      |      timestamp : value                     |
//      |      timestamp : value                     |
//      |      timestamp : value                     |
//      +--------------------------------------------+
//      |                                            |
//
//  We then insert each sample multiple times, with different length prefixes:
//
//      TIME_PREFIX      EXAMPLE                         MEANING
//      YY               83f0a...|temperature|15         Year worth of samples
//      YYMM             83f0a...|temperature|1503       Month worth of samples
//      YYMMDD           83f0a...|temperature|150314     Day worth of samples
//      YYMMDDHH         83f0a...|temperature|15031403   Hour worth of samples
//
//  For weekly data, we specify the first day (monday) of the week and append
//  the letter "w":
//
//      YYMMDDw         83f0a...|temperature|150309w    Week worth of samples
//
//  Our software ensures that each bucket contains a reasonable number of
//  samples.  For example, a YY bucket containing a Year's worth of samples may
//  only have 1 sample/day, whereas a YYMMDD bucket containing a Day's worth of
//  samples may have a sample every 5 minutes.  We can easily trim away old
//  samples by deleting the appropriate rows.
//
//  You can gain insight into the actual structure of a CF by running:
//
//      > cassandra-cli
//      > use canopy;
//      > list propval_float;
//
//  Also useful:
//      > nodetool cfstats
//
//

/* Very useful: http://www.datastax.com/dev/blog/thrift-to-cql3 */
var creationQueries []string = []string{
    // Keeps track of last update time for cloud variable
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
    //  void
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

    // used for:
    //  uint8
    //  int8
    //  int16
    //  uint16
    //  int32
    //  uint32
    `CREATE TABLE propval_int (
        device_id uuid,
        propname text,
        time timestamp,
        value int,
        PRIMARY KEY((device_id, propname), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  float32
    `CREATE TABLE propval_float (
        device_id uuid,
        propname text,
        time timestamp,
        value float,
        PRIMARY KEY((device_id, propname), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  float64
    `CREATE TABLE propval_double (
        device_id uuid,
        propname text,
        time timestamp,
        value double,
        PRIMARY KEY((device_id, propname), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  datetime
    `CREATE TABLE propval_timestamp (
        device_id uuid,
        propname text,
        time timestamp,
        value timestamp,
        PRIMARY KEY((device_id, propname), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  bool
    `CREATE TABLE propval_boolean (
        device_id uuid,
        propname text,
        time timestamp,
        value boolean,
        PRIMARY KEY((device_id, propname), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  void
    `CREATE TABLE propval_void (
        device_id uuid,
        propname text,
        time timestamp,
        PRIMARY KEY((device_id, propname), time)
    ) WITH COMPACT STORAGE`,

    // used for:
    //  string
    `CREATE TABLE propval_string (
        device_id uuid,
        propname text,
        time timestamp,
        value text,
        PRIMARY KEY((device_id, propname), time)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE var_sample_counts (
        device_id uuid,
        vardecl text,
        count counter,
        PRIMARY KEY(device_id, vardecl)
    )`,
    // var_info
    // This table stores config and state relating to cloud variables.
    //  device_id
    //      Device that owns the cloud variable.
    //
    //  vardecl
    //      Cloud variable declaration, such as "inout float32 humidity".
    //      The combinatino of (deviceid, vardecl) uniquely identifies the
    //      cloud variable.
    //
    //  sample_limit
    //      Maximum number of samples to keep until we start discarding.
    `CREATE TABLE var_info (
        device_id uuid,
        vardecl text,
        sample_limit int,
        PRIMARY KEY(device_id, vardecl)
    )`,

    `CREATE TABLE devices (
        device_id uuid,
        secret_key text,
        friendly_name text,
        location_note text,
        sddl text,
        public_access_level int,
        last_seen timestamp,
        ws_connected boolean,
        PRIMARY KEY(device_id)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE device_group (
        username text,
        group_name text,
        group_order int,
        device_id uuid,
        device_friendly_name text,
        PRIMARY KEY(username, group_name, group_order)
    )`,

    `CREATE TABLE control_event (
        device_id uuid,
        time_issued timestamp,
        control_name text,
        value double,
        PRIMARY KEY(device_id, time_issued)
    )`,

    `CREATE TABLE device_permissions (
        username text,
        device_id uuid,
        access_level int,
        PRIMARY KEY(username, device_id)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE accounts (
        username text,
        email text,
        password_hash blob,
        activated boolean,
        activation_code text,
        password_reset_code text,
        password_reset_code_expiry timestamp,
        PRIMARY KEY(username)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE account_emails (
        email text,
        username text,
        PRIMARY KEY(email)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE notifications (
        device_id uuid,
        time_issued timestamp,
        dismissed boolean,
        msg text,
        notify_type int,
        PRIMARY KEY(device_id, time_issued)
    ) `,

    `CREATE TABLE workers (
        name text,
        status text,
        PRIMARY KEY(name)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE listeners (
        key text
        workers set<text>,
        PRIMARY KEY(key)
    ) `,
}

type CassDatalayer struct {
    cfg config.Config
}

func NewCassDatalayer(cfg config.Config) *CassDatalayer {
    return &CassDatalayer{cfg: cfg}
}

func (dl *CassDatalayer) Connect(keyspace string) (datalayer.Connection, error) {
    cluster := gocql.NewCluster("127.0.0.1")
    cluster.Keyspace = keyspace
    cluster.Consistency = gocql.Any

    session, err := cluster.CreateSession()
    if err != nil {
        canolog.Error("Error creating DB session: ", err)
        return nil, err
    }

    return &CassConnection{
        dl: dl,
        session: session,
    }, nil
}

func (dl *CassDatalayer) EraseDb(keyspace string) error {
    cluster := gocql.NewCluster("127.0.0.1")

    session, err := cluster.CreateSession()
    if err != nil {
        canolog.Error("Error creating DB session: ", err)
        return err
    }

    err = session.Query(`DROP KEYSPACE ` + keyspace + ``).Exec()
    return err
}

func (dl *CassDatalayer) PrepDb(keyspace string) error {
    cluster := gocql.NewCluster("127.0.0.1")

    session, err := cluster.CreateSession()
    if err != nil {
        canolog.Error("Error creating DB session: ", err)
        return err
    }

    // Create keyspace.
    err = session.Query(`
            CREATE KEYSPACE ` + keyspace + `
            WITH REPLICATION = {'class' : 'SimpleStrategy', 'replication_factor' : 1}
    `).Exec()
    if err != nil {
        // Ignore errors (just log them).
        canolog.Warn("(IGNORED) ", err)
    }

    // Create a new session connecting to that keyspace.
    cluster = gocql.NewCluster("127.0.0.1")
    cluster.Keyspace = keyspace
    cluster.Consistency = gocql.Quorum
    session, err = cluster.CreateSession()
    if err != nil {
        canolog.Error("Error creating DB session: ", err)
        return err
    }

    // Perform all creation queries.
    for _, query := range creationQueries {
        if err := session.Query(query).Exec(); err != nil {
            // Ignore errors (just print them).
            // This allows PrepDB to be used to add new tables.  Eventually, we
            // should come up with a proper migration strategy.
            canolog.Warn("(IGNORED) ", query, ": ", err)
        }
    }
    return nil
}

// Migrate to next version of database
// Returns version of DB after migration
func (dl *CassDatalayer) migrateNext(session *gocql.Session, startVersion string) (string, error) {
    if startVersion == "0.9.0" {
        err := migrations.Migrate_0_9_0_to_0_9_1(session)
        if err != nil {
            return startVersion, err
        }
        return "0.9.1", nil
    }
    return  startVersion, fmt.Errorf("Unknown DB version %s", startVersion)
}

func (dl *CassDatalayer) MigrateDB(keyspace, startVersion, endVersion string) error {
    var err error
    cluster := gocql.NewCluster("127.0.0.1")
    cluster.Keyspace = keyspace

    session, err := cluster.CreateSession()
    if err != nil {
        canolog.Error("Error creating DB session: ", err)
        return err
    }

    curVersion := startVersion
    for curVersion != endVersion {
        canolog.Info("Migrating from %s to next version", curVersion)
        curVersion, err = dl.migrateNext(session, startVersion)
        if err != nil {
            canolog.Error("Failed migrating from %s:", curVersion, err)
            return err
        }
    }
    canolog.Info("Migration complete!  DB is now version: %s", curVersion)
    return nil
}

func NewDatalayer(cfg config.Config) datalayer.Datalayer {
    return NewCassDatalayer(cfg)
}
