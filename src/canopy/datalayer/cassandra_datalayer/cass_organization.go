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
    "canopy/canolog"
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
    // Add account to team
    err := org.conn.session.Query(`
            INSERT INTO account_teams (username, org_id, name)
            VALUES (?, ?, ?)
    `, account.Username(), org.id, team).Exec()
    if err != nil {
        canolog.Error("Error adding account to team: ", err)
        return err
    }

    // Add account to organization member table
    // TODO: sanitize inputs
    err = org.AddMember(account)
    return err
}

func (org *CassOrganization) AddMember(account datalayer.Account) error {
    // Add account to organization member table
    // TODO: sanitize inputs
    err := org.conn.session.Query(`
            UPDATE organization_members
            SET usernames = usernames + {'` + account.Username() + `'}
            WHERE name = ?
    `, org.name).Exec()
    if err != nil {
        canolog.Error("Error adding account as member of organization: ", err)
        return err;
    }
    return nil
}

func (org *CassOrganization) CreateTeam(team string) error {
    // Create Team
    err := org.conn.session.Query(`
            INSERT INTO teams (org_id, name)
            VALUES (?, ?)
    `, org.id, team).Exec()
    if err != nil {
        canolog.Error("Error storing team: ", err)
        return err
    }

    return nil
}

func (org *CassOrganization) ID() string {
    return org.id.String()
}

func (org *CassOrganization) IsMember(account datalayer.Account) (bool, error) {
    // TODO: inefficient
    members, err := org.Members()
    if err != nil {
        return false, err
    }

    for _, member := range members {
        if member.Username() == account.Username() {
            return true, nil
        }
    }

    return false, nil
}

func (org *CassOrganization) IsOwner(account datalayer.Account) (bool, error) {
    return false, fmt.Errorf("Not implemented")
}

func (org *CassOrganization) Members() ([]datalayer.Account, error) {
    var usernames []string
    rows, err := org.conn.session.Query(`
            SELECT usernames FROM organization_members
            WHERE name = ?
    `, org.name).Consistency(gocql.One).Iter().SliceMap();
    if err != nil {
        canolog.Error(err)
    }
    if len(rows) != 1 {
        return nil, fmt.Errorf("Expected 1 DB row for usernames ")
    }
    usernames = rows[0]["usernames"].([]string)

    // Lookup account objects
    // TODO: inefficient manual join here
    out := []datalayer.Account{}
    for _, username := range usernames {
        account, err := org.conn.LookupAccount(username)
        if err != nil {
            return []datalayer.Account{}, err
        }
        out = append(out, account)
    }

    return out, nil
}

func (org *CassOrganization) Name() string {
    return org.name
}

func (org *CassOrganization) RemoveMember(account datalayer.Account) error {
    // Remove account from organization member table
    // TODO: sanitize inputs
    err := org.conn.session.Query(`
            UPDATE organization_members
            SET usernames = usernames - {'` + account.Username() + `'}
            WHERE name = ?
    `, org.name).Exec()
    if err != nil {
        canolog.Error("Error removing account as member of organization: ", err)
        return err;
    }

    // TODO: Remove from teams table
    //
    // TODO: Remove from account membership table
    return fmt.Errorf("Not fully implemented")
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
