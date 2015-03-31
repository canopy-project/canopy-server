/*
 * Copright 2014-2015 Canopy Services, Inc.
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
    "fmt"
    "github.com/gocql/gocql"
)

type CassPigeonSystem struct {
    conn *CassConnection
}

func (pigeonsys *CassPigeonSystem) GetListeners(key string) ([]string, error) {
    var workers []string
    rows, err := pigeonsys.conn.session.Query(`
            SELECT * FROM listeners
            WHERE key = ?
    `, key).Consistency(gocql.One).Iter().SliceMap();
    if err != nil {
        canolog.Error(err)
    }
    if len(rows) != 1 {
        return nil, fmt.Errorf("Expected 1 DB row for listener ", key)
    }
    workers = rows[0]["workers"].([]string)
    return workers, nil
}

func (pigeonsys *CassPigeonSystem) RegisterListener(hostname, key string) error {
    // TODO: sanitize inputs
    err := pigeonsys.conn.session.Query(`
            UPDATE listeners
            SET workers = workers + {'` + hostname + `'}
            WHERE key = ?
    `, key).Exec()
    if err != nil {
        return err;
    }
    return nil
}

func (pigeonsys *CassPigeonSystem) RegisterWorker(hostname string) error {
    err := pigeonsys.conn.session.Query(`
            UPDATE workers
            SET status = ?
            WHERE name = ?
    `, "A", hostname).Exec()
    if err != nil {
        return err;
    }
    return nil
}
