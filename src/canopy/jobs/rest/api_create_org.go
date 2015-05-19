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
    "canopy/datalayer"
)

// Constructs the response body for the /api/create_org REST endpoint
func ApiCreateOrgHandler(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    if info.Account == nil {
        return nil, UnauthorizedError("User credentials required to create Organization").Log()
    }
    name, ok := info.BodyObj["name"].(string)
    if !ok {
        return nil, BadInputError("String \"name\" expected").Log()
    }

    org, err := info.Account.CreateOrganization(name)
    if err != nil {
        switch err.(type) {
        case *datalayer.ValidationError:
            return nil, BadInputError(err.Error()).Log()
        default:
            return nil, InternalServerError("Problem Creating Organization: " + err.Error()).Log()
        }
    }

    out := map[string]interface{} {
        "name" : org.Name(),
        "result" : "ok",
    }
    return out, nil
}
