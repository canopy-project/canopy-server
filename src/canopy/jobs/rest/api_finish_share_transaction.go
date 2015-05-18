// Copyright 2014-2015 Canopy Services, Inc.
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

func POST__api__finish_share_transaction(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    /*
     *  POST
     *  {
     *      "device_id" : <DEVICE_ID>,
     *  }
     *
     * TODO: Add to REST API documentation
     * TODO: Highly insecure!!!
     */
    deviceId, ok := info.BodyObj["device_id"].(string)
    if !ok {
        return nil, BadInputError("String \"device_id\" expected")
    }

    if info.Account == nil {
        return nil, NotLoggedInError()
    }

    device, err := info.Conn.LookupDevice(deviceId)
    if err != nil {
        // TODO: return proper error
        return nil, InternalServerError("Looking up device")
    }

    /* Grant permissions to the user to access the device */
    err = device.SetAccountAccess(info.Account, datalayer.ReadWriteAccess, datalayer.ShareRevokeAllowed)
    if err != nil {
        return nil, InternalServerError("Could not grant access")
    }

    return map[string]interface{} {
        "result" : "ok",
        "device_friendly_name" : device.Name(),
    }, nil
}
