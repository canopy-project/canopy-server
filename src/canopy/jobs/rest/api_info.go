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
    "canopy/config"
    "canopy/jobqueue"
    "fmt"
)

// Constructs the response body for the /api/info REST endpoint
func ApiInfoHandler(userCtx map[string]interface{}, req jobqueue.Request, resp jobqueue.Response) {

    cfg, ok := userCtx["cfg"].(config.Config)
    if !ok {
        resp.SetError(fmt.Errorf("Internal error: expected 'cfg' in UserContext"))
        return
    }

    resp.SetBody(map[string]interface{}{
        "result" : "ok",
        "service-name" : "Canopy Cloud Service",
        "version" : "0.9.2-beta",
        "config" : cfg.ToJsonObject(),
    })
}
