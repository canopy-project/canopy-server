// Copyright 2014 SimpleThings, Inc.
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
package endpoints

import (
    "canopy/datalayer"
    "canopy/rest/adapter"
    "canopy/rest/rest_errors"
    "net/http"
)

func POST_create_devices(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
    if info.Account == nil {
        return nil, rest_errors.NewNotLoggedInError()
    }

    quantityFloat, ok := info.BodyObj["quantity"].(float64)
    if !ok {
        return nil, rest_errors.NewBadInputError("Numeric \"quantity\" expected")
    }
    quantity := int(quantityFloat)

    friendlyNames, ok := info.BodyObj["friendly_names"].([]interface{})
    if !ok {
        return nil, rest_errors.NewBadInputError("List \"friendly_names\" expected")
    }

    if len(friendlyNames) != quantity {
        return nil, rest_errors.NewBadInputError("Incorrect number of friendly_names provided")
    }

    out := map[string]interface{} {
        "result" : "ok",
        "devices" : []interface{} {},
    }

    for _, nameItf := range friendlyNames {
        friendlyName, ok := nameItf.(string)
        if !ok {
            return nil, rest_errors.NewBadInputError("String friendly name expected")
        }
        device, err := info.Conn.CreateDevice(friendlyName, nil, "", datalayer.NoAccess);
        if err != nil {
            return nil, rest_errors.NewInternalServerError("Error creating device")
        }

        err = device.SetAccountAccess(info.Account, datalayer.ReadWriteAccess, datalayer.ShareRevokeAllowed);
        if err != nil {
            return nil, rest_errors.NewInternalServerError("Error setting device permissions")
        }

        devicesSlice, ok := out["devices"].([]interface{})
        out["devices"] = append(devicesSlice, map[string]interface{} {
            "friendly_name" : device.Name(),
            "device_id" : device.ID(),
            "device_secret_key" : device.SecretKey(),
        })
    }

    return out, nil
}
