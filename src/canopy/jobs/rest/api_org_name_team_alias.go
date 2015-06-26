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

// Backend implementation DELETE /api/org/{name}/team/{alias} endpoint
// 
func DELETE__api__org__name__team__alias(info *RestRequestInfo, sideEffects *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, NotLoggedInError().Log()
    }

    // Lookup organization by name
    org, err := info.Conn.LookupOrganization(info.URLVars["name"])
    if err != nil {
        // TODO: determine actual cause of error
        return nil, URLNotFoundError().Log()
    }

    // Is authenticated user owner of this organization?
    ok, err := org.IsOwner(info.Account)
    if !ok || err != nil {
        return nil, UnauthorizedError("Must be owner of org to delete team").Log()
    }

    alias := info.URLVars["alias"]

    // Delete team
    err = org.DeleteTeam(alias)
    if err != nil {
        // TODO: determine actual cause of error
        return nil, InternalServerError(err.Error()).Log()
    }

    out := map[string]interface{} {
        "result" : "ok",
    }

    return out, nil
}
