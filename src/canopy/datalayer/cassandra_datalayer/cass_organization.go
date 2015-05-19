/*
 * Copyright 2014-2015 Canopy Services, Inc.
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
    "fmt"
    "github.com/gocql/gocql"
)

type CassOrganization struct {
    conn *CassConnection
    id gocql.UUID
    name string
}

func (org *CassOrganization) AddAccountToTeam(account datalayer.Account, team string) error {
    return fmt.Errorf("not implemented yet")
}

func (org *CassOrganization) CreateTeam(team string) error {
    // Create Team
    return fmt.Errorf("not implemented yet")
}

func (org *CassOrganization) ID() string {
    return org.id.String()
}

func (org *CassOrganization) Name() string {
    return org.name
}

func (org *CassOrganization) SetName(name string) error {
    available, err := org.conn.IsNameAvailable(name)
    if err != nil {
        return fmt.Errorf("Internal error occurred check for name availability")
    } else if !available{
        return fmt.Errorf("That name is not available")
    }

    err = org.conn.session.Query(`
            UPDATE organization_names
            SET name = ?,
            WHERE id = ?
    `, name, org.id).Exec()
    if err != nil {
        return err;
    }

    org.name = name
    return nil
}
