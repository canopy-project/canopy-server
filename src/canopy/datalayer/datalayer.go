package datalayer

/* Very useful: http://www.datastax.com/dev/blog/thrift-to-cql3 */
import (
    "github.com/gocql/gocql"
    "log"
)
var creationQueries []string = []string{
    `CREATE TABLE sensor_data (
        device_id uuid,
        propname text,
        time timestamp,
        value double,
        PRIMARY KEY(device_id, propname, time)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE devices (
        device_id uuid,
        friendly_name text,
        sddl text,
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
        PRIMARY KEY(username)
    ) WITH COMPACT STORAGE`,

    `CREATE TABLE account_emails (
        email text,
        username text,
        PRIMARY KEY(email)
    ) WITH COMPACT STORAGE`,
}

type CassandraDatalayer struct {
    
    cluster *gocql.ClusterConfig
    session *gocql.Session
}

func NewCassandraDatalayer() *CassandraDatalayer {
    return &CassandraDatalayer{cluster: nil, session: nil}
}

func (dl *CassandraDatalayer) Connect(keyspace string) {
    dl.cluster = gocql.NewCluster("127.0.0.1")
    dl.cluster.Keyspace = keyspace
    dl.cluster.Consistency = gocql.Any
    dl.session, _ = dl.cluster.CreateSession()
}

func (dl *CassandraDatalayer) Close() {
    dl.session.Close()
}

func (dl *CassandraDatalayer) StorePropertyValue(device_id string, propname string, value float64) {
    /* deprecated */
    if err := dl.session.Query(`
            INSERT INTO sensor_data (device_id, propname, time, value)
            VALUES (?, ?, dateof(now()), ?)
    `, device_id, propname, value).Exec(); err != nil {
        log.Print(err)
    }
}

func (dl *CassandraDatalayer) EraseDb(keyspace string) {
    dl.cluster = gocql.NewCluster("127.0.0.1")
    dl.session, _ = dl.cluster.CreateSession()
    if err := dl.session.Query(`
        DROP KEYSPACE ` + keyspace + `
    `).Exec(); err != nil {
        log.Print(err)
    }
}

func (dl *CassandraDatalayer) PrepDb(keyspace string) {
    dl.cluster = gocql.NewCluster("127.0.0.1")
    dl.session, _ = dl.cluster.CreateSession()
    if err := dl.session.Query(`
            CREATE KEYSPACE ` + keyspace + `
            WITH REPLICATION = {'class' : 'SimpleStrategy', 'replication_factor' : 3}
    `).Exec(); err != nil {
        log.Print(err)
    }

    dl.cluster = gocql.NewCluster("127.0.0.1")
    dl.cluster.Keyspace = keyspace
    dl.cluster.Consistency = gocql.Quorum
    dl.session, _ = dl.cluster.CreateSession()

    for _, query := range creationQueries {
        if err := dl.session.Query(query).Exec(); err != nil {
            log.Print(query, "\n", err)
        }
    }
}
