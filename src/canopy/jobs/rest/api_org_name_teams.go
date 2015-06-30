// Copyright 2015 Canopy Services, Inc.
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

package rest

import (
)

// Backend implementation GET /api/org/{name}/teams endpoint
// 
func GET__api__org__name__teams(info *RestRequestInfo, sideEffects *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, NotLoggedInError().Log()
    }

    // Lookup organization by name
    org, err := info.Conn.LookupOrganization(info.URLVars["name"])
    if err != nil {
        // TODO: determine actual cause of error
        return nil, URLNotFoundError().Log()
    }

    // Is authenticated user member of this organization?
    ok, err := org.IsMember(info.Account)
    if !ok || err != nil {
        return nil, UnauthorizedError("Must be member of org to see teams").Log()
    }

    // Get teams list
    teams, err := org.Teams()
    if err != nil {
        return nil, InternalServerError(err.Error()).Log()
    }

    // Generate response payload
    teamsJsonObj := []interface{}{}
    for _, team := range teams {
        teamsJsonObj = append(teamsJsonObj, map[string]string {
            "name" : team.Name(),
            "url_alias" : team.UrlAlias(),
        })
    }

    out := map[string]interface{} {
        "result" : "ok",
        "teams" : teamsJsonObj,
    }

    return out, nil
}
