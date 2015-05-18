/*
 * Copright 2014-2015 Canopy Services, Inc.
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
    "canopy/cloudvar"
    "canopy/datalayer"
    "canopy/device_filter"
    "github.com/gocql/gocql"
    "sort"
)

type CassDeviceQuery struct {
    account *CassAccount
    sortOrder []string
    filterExpr string
}

func (dq *CassDeviceQuery)Copy() *CassDeviceQuery {
    out := &CassDeviceQuery{}
    *out = *dq
    return out
}

func (dq *CassDeviceQuery)Filter(expr string) datalayer.DeviceQuery {
    out := dq.Copy()
    out.filterExpr = expr
    return out
}

func (dq *CassDeviceQuery)SortBy(order ...string) datalayer.DeviceQuery {
    out := dq.Copy()
    out.sortOrder = order
    return out
}

type sortData struct {
    devices []datalayer.Device
    sortOrder []string
}

func (data sortData) Len() int {
    return len(data.devices)
}
func (data sortData) Swap(i, j int) {
    data.devices[i], data.devices[j] = data.devices[j], data.devices[i]
}
func (data sortData) Less(i, j int) bool {
    var s int
    for s = 0; s < len(data.sortOrder); s++ {
        descending := ([]rune(data.sortOrder[s])[0] == '-')
        var sortKey string
        if descending {
            sortKey = data.sortOrder[s][1:]
        } else {
            sortKey = data.sortOrder[s]
        }
        varDefA, errA := data.devices[i].LookupVarDef(sortKey)
        varDefB, errB := data.devices[j].LookupVarDef(sortKey)

        if errA != nil && errB != nil {
            continue
        } else if errA != nil && errB == nil{
            return !descending
        } else if errA == nil && errB != nil {
            return descending
        }

        sampleA, errA := data.devices[i].LatestData(varDefA)
        sampleB, errB := data.devices[j].LatestData(varDefB)

        if errA != nil && errB != nil {
            continue
        } else if errA != nil {
            return !descending
        } else if errB != nil {
            return descending
        }

        // TOOD: support descending
        // TODO: support secondary, tertiary, etc
        // TODO: What happens if datatype differs?
        less, _ := cloudvar.Less(varDefA.Datatype(), sampleA.Value, sampleB.Value)
        if less {
            return !descending
        }
        greater, _ := cloudvar.Greater(varDefA.Datatype(), sampleA.Value, sampleB.Value)
        if greater {
            return descending
        }
    }

    // Tie breaker: Device name
    if data.devices[i].Name() < data.devices[j].Name() {
        return true
    } else if data.devices[j].Name() < data.devices[i].Name() {
        return false
    }

    // Ultimate tie breaker: Device UUID
    if (data.devices[i].ID() < data.devices[j].ID()) {
        return true
    }
    return false
}

func (dq *CassDeviceQuery)DeviceList(start, count int32) ([]datalayer.Device, error) {
    devices := []datalayer.Device{}
    var deviceId string
    var accessLevel int

    // Fetch all devices (TODO: inefficient!)
    query := dq.account.conn.session.Query(`
            SELECT device_id, access_level FROM device_permissions_v2
            WHERE username = ?
    `, dq.account.Username()).Consistency(gocql.One)
    iter := query.Iter()
    for iter.Scan(&deviceId, &accessLevel) {
        if accessLevel > 0 {
            device, err := dq.account.conn.LookupDevice(deviceId)
            if err != nil {
                iter.Close()
                return []datalayer.Device{}, err
            }
            devices = append(devices, device)
        }
    }
    if err := iter.Close(); err != nil {
        return []datalayer.Device{}, err
    }

    // Filter
    if dq.filterExpr != "" {
        filter, err := device_filter.Compile(dq.filterExpr)
        if err != nil {
            return []datalayer.Device{}, err
        }

        devices, err = filter.Whittle(devices)
        if err != nil {
            return []datalayer.Device{}, err
        }
    }

    // Sort
    var data sortData
    if dq.sortOrder != nil {
        data = sortData{ devices: devices, sortOrder: dq.sortOrder }
    } else {
        data = sortData{ devices: devices, sortOrder: []string{}}
    }
    sort.Sort(data)

    // Apply limits
    out := []datalayer.Device{}
    var i int32
    if start < 0 {
        start = 0
    }
    if count == -1 {
        count = int32(len(devices))
    }
    end := start + count
    if end > int32(len(devices)) {
        end = int32(len(devices))
    }

    for i = start; i < end; i++ {
        out = append(out, devices[i])
    }

    return out, nil
}

func (dq *CassDeviceQuery)Count() (int32, error) {
    // TODO: Inefficient
    devices, err := dq.DeviceList(0, -1)
    if err != nil {
        return 0, err
    }
    return int32(len(devices)), nil
}
