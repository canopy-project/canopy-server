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

type CassTeam struct {
    org *CassOrganization
    urlAlias string
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
    //err = org.AddMember(account, false)
    return err
}

func (org *CassOrganization) AddMember(account datalayer.Account, owner bool) error {
    // TODO: transactionize
 
    // Add account to organization_membership
    err := org.conn.session.Query(`
            INSERT INTO organization_membership
                (org_id, username, is_owner)
            VALUES
                (?, ?, ?)
    `, org.id, account.Username(), owner).Exec()
    if err != nil {
        canolog.Error("Error adding account as member of organization: ", err)
        return err;
    }

    // Add account to account_orgs (this table acts as an index)
    err = org.conn.session.Query(`
            INSERT INTO account_orgs
                (username, org_id)
            VALUES
                (?, ?)
    `, account.Username(), org.id).Exec()
    if err != nil {
        canolog.Error("Error indexing account as member of organization: ", err)
        return err;
    }

    return nil
}

func (org *CassOrganization) CreateTeam(team string, url_alias string) error {
    // TODO: validate input
    // TODO: URLify
    // TODO: Check if team already exists

    // Create Team
    err := org.conn.session.Query(`
            INSERT INTO teams (org_id, url_alias, name)
            VALUES (?, ?, ?)
    `, org.id, url_alias, team).Exec()
    if err != nil {
        canolog.Error("Error storing team: ", err)
        return err
    }

    return nil
}

func (org *CassOrganization) DeleteTeam(url_alias string) error {
    // TODO: Check if team already exists

    // Delete Team
    err := org.conn.session.Query(`
            DELETE FROM teams
            WHERE org_id = ?  AND url_alias = ?
    `, org.id, url_alias).Exec()
    if err != nil {
        canolog.Error("Error deleting team from DB: ", err)
        return err
    }

    // Delete Team Membership
    err = org.conn.session.Query(`
            DELETE FROM team_membership
            WHERE org_id = ?  AND team_url_alias = ?
    `, org.id, url_alias).Exec()
    if err != nil {
        canolog.Error("Error deleting team membership data from DB: ", err)
        return err
    }

    // TODO: also delete account_membership data

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
        if member.Account.Username() == account.Username() {
            return true, nil
        }
    }

    return false, nil
}

func (org *CassOrganization) IsOwner(account datalayer.Account) (bool, error) {
    // TODO: inefficient
    members, err := org.Members()
    if err != nil {
        return false, err
    }

    for _, member := range members {
        if (member.Account.Username() == account.Username()) && member.IsOwner {
            return true, nil
        }
    }

    return false, nil
}

func (org *CassOrganization) Members() ([]datalayer.OrganizationMemberInfo, error) {
    var out []datalayer.OrganizationMemberInfo
    rows, err := org.conn.session.Query(`
            SELECT username, is_owner 
            FROM organization_membership
            WHERE org_id = ?
    `, org.id).Consistency(gocql.One).Iter().SliceMap();
    if err != nil {
        canolog.Error(err)
        return []datalayer.OrganizationMemberInfo{}, err
    }
    for _, row := range rows {
        // TODO: inefficient manual join here
        username := row["username"].(string)
        isOwner := row["is_owner"].(bool)
        account, err := org.conn.LookupAccount(username)
        if err != nil {
            return []datalayer.OrganizationMemberInfo{}, err
        }
        out = append(out, datalayer.OrganizationMemberInfo{account, isOwner})
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
            DELETE FROM organization_membership
            WHERE org_id = ?  AND username = ?
    `, org.id, account.Username()).Exec()
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

func (org *CassOrganization) Team(teamUrlAlias string) (datalayer.Team, error) {
    var name string

    err := org.conn.session.Query(`
        SELECT name FROM teams
        WHERE org_id = ? AND url_alias = ?
        LIMIT 1
    `, org.id, teamUrlAlias).Consistency(gocql.One).Scan(&name)
    if err != nil {
        return nil, err
    }

    // Create team object
    team := &CassTeam{
        org: org,
        urlAlias: teamUrlAlias,
        name: name,
    }

    return team, nil
}

func (org *CassOrganization) Teams() ([]datalayer.Team, error) {
    var out []datalayer.Team

    rows, err := org.conn.session.Query(`
        SELECT team_url_alias, name FROM teams
        WHERE org_id = ?
        `, org.id).Consistency(gocql.One).Iter().SliceMap();
    if err != nil {
        canolog.Error(err)
        return []datalayer.Team{}, err
    }
    for _, row := range rows {
        // Create team object
        team := &CassTeam{
            org: org,
            urlAlias: row["url_alias"].(string),
            name: row["name"].(string),
        }
        out = append(out, team)
    }
    return out, nil
}

func (team *CassTeam) AddMember(account datalayer.Account) error {
    // Is account a member of the organization?
    isMember, err := team.org.IsMember(account)
    if !isMember {
        return fmt.Errorf("Only organization members can be added to team")
    } else if err != nil {
        return err
    }

    // Add to team_membership table
    err = team.org.conn.session.Query(`
            INSERT INTO team_membership (org_id, team_url_alias, username)
            VALUES (?, ?, ?)
    `, team.org.id, team.urlAlias, account.Username()).Exec()
    if err != nil {
        canolog.Error("Error adding account as member of team: ", err)
        return err;
    }

    // TODO: Add to account_teams table?
    /*err := org.conn.session.Query(
            INSERT INTO account_teams (org_id, team_url_alias, username)
            VALUES (?, ?, ?)
    `, org.id, team.urlAlias, account.Username()).Exec()
    if err != nil {
        canolog.Error("Error adding account as member of team: ", err)
        return err;*/

    return nil
}

func (team *CassTeam) Name() string {
    return team.name
}

func (team *CassTeam) UrlAlias() string {
    return team.urlAlias
}

func (team *CassTeam) Members() ([]datalayer.OrganizationMemberInfo, error) {
    var out []datalayer.OrganizationMemberInfo
    rows, err := team.org.conn.session.Query(`
            SELECT username, is_owner 
            FROM team_membership
            WHERE org_id = ? AND team_url_alias = ?
    `, team.org.id, team.urlAlias).Consistency(gocql.One).Iter().SliceMap();
    if err != nil {
        canolog.Error(err)
        return []datalayer.OrganizationMemberInfo{}, err
    }
    for _, row := range rows {
        // TODO: inefficient manual join here
        username := row["username"].(string)
        isOwner := row["is_owner"].(bool)
        account, err := team.org.conn.LookupAccount(username)
        if err != nil {
            return []datalayer.OrganizationMemberInfo{}, err
        }
        out = append(out, datalayer.OrganizationMemberInfo{account, isOwner})
    }
    return out, nil
}

func (team *CassTeam) RemoveMember(account datalayer.Account) error {
    // Remove from team_membership table
    err := team.org.conn.session.Query(`
            DELETE FROM team_membership
            WHERE org_id = ?  AND team_url_alias = ? AND username = ?
    `, team.org.id, team.urlAlias, account.Username()).Exec()
    if err != nil {
        canolog.Error("Error removing account as member of team: ", err)
        return err;
    }

    return nil
}
