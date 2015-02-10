// Copyright 2014-2015 SimpleThings, Inc.
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

var migrationQueries []string = []string{
    // Add var_sample_counts table
    `CREATE TABLE var_sample_counts (
        device_id uuid,
        vardecl text,
        count counter,
        PRIMARY KEY(device_id, vardecl)
    )`,

    // Add var_info table
    `CREATE TABLE var_info (
        device_id uuid,
        vardecl text,
        sample_limit int,
        PRIMARY KEY(device_id, vardecl)
    )`,

    // Add password_reset_code and password_reset_code_expiry to accounts
    `ALTER TABLE accounts ADD password_reset_code text`,
    `ALTER TABLE accounts ADD password_reset_code_expiry timestamp`,
}
func Migrate_0_9_0_to_0_9_1(session *gocql.Session) error {
    // Perform all migration queries.
    for _, query := range migrationQueries {
        if err := session.Query(query).Exec(); err != nil {
            // Ignore errors (just print them).
            canolog.Warn(query, ": ", err)
            return err
        }
    }
    return nil
}
