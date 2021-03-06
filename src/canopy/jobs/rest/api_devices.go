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
package rest

import (
    "strconv"
    "strings"
)

func GET__api__devices(info *RestRequestInfo, sideEffects *RestSideEffects) (map[string]interface{}, RestError) {
    var err error

    if info.Account == nil {
        return nil, NotLoggedInError()
    }

    dq := info.Account.Devices()

    // Filter
    filterExpr := info.Query["filter"]
    if filterExpr != nil {
        dq = dq.Filter(filterExpr[0])
    }

    // Paging
    totalCount, err := dq.Count()
    if err != nil {
        return nil, InternalServerError("Error determining device count: " + err.Error())
    }

    limit := info.Query["limit"]
    start := int64(0)
    count := int64(-1)
    if limit != nil {
        limitStrings := strings.Split(limit[0], ",")
        if len(limitStrings) != 2 {
            return nil, BadInputError("Expected \"start,count\" for \"limit\"")
        }
        start, err = strconv.ParseInt(limitStrings[0], 10, 32)
        if err != nil {
            return nil, BadInputError("Expected int for limit start")
        }
        count, err = strconv.ParseInt(limitStrings[1], 10, 32)
        if err != nil {
            return nil, BadInputError("Expected int for limit count")
        }
    }

    sort := info.Query["sort"]
    if sort != nil {
        sortStrings := strings.Split(sort[0], ",")
        dq = dq.SortBy(sortStrings...)
    }

    devices, err := dq.DeviceList(int32(start), int32(count))
    if err != nil {
        return nil, InternalServerError("Device lookup failed")
    }

    timestamps := info.Query["timestamps"]
    timestamp_type := "epoch_us"
    if timestamps != nil && timestamps[0] == "rfc3339" {
        timestamp_type = "rfc3339"
    }

    //out, err := devicesToJsonObj(info.PigeonSys, devices)
    // TODO: How do we tell ws connectivity status?
    out, err := devicesToJsonObj(devices, timestamp_type)
    if err != nil {
        return nil, InternalServerError("Generating JSON")
    }
    out["result"] = "ok"
    out["paging"] = map[string]interface{}{
        "total_count" : totalCount,
    }

    return out, nil
}
