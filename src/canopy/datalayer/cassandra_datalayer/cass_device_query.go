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
    "github.com/gocql/gocql"
    "fmt"
    "sort"
)

type CassDeviceQuery struct {
    account *CassAccount
    sortOrder []string
    limitStart int32
    limitCount int32
    filters map[string]interface{}
}

func (dq *CassDeviceQuery)Copy() *CassDeviceQuery {
    out := &CassDeviceQuery{}
    *out = *dq
    return out
}

func (dq *CassDeviceQuery)SetFilter(expr string) (datalayer.DeviceQuery, error) {
    return nil, fmt.Errorf("Filter not yet implemented")
}

func (dq *CassDeviceQuery)SetSortOrder(order ...string) (datalayer.DeviceQuery, error) {
    out := dq.Copy()
    out.sortOrder = order
    return out, nil
}

func (dq *CassDeviceQuery)SetLimits(start, count int32) (datalayer.DeviceQuery, error) {
    out := dq.Copy()
    out.limitStart = start
    out.limitCount = count
    return out, nil
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
    varDefA, errA := data.devices[i].LookupVarDef(data.sortOrder[0])
    varDefB, errB := data.devices[j].LookupVarDef(data.sortOrder[0])

    if errA != nil && errB != nil {
        return data.devices[i].IDString() < data.devices[j].IDString()
    } else if errA != nil {
        return false
    } else if errB != nil {
        return true
    }

    sampleA, errA := data.devices[i].LatestData(varDefA)
    sampleB, errB := data.devices[j].LatestData(varDefB)

    if errA != nil && errB != nil {
        return data.devices[i].IDString() < data.devices[j].IDString()
    } else if errA != nil {
        return false
    } else if errB != nil {
        return true
    }

    // TOOD: support descending
    // TODO: support secondary, tertiary, etc
    // TODO: What happens if datatype differs?
    less, _ := cloudvar.Less(varDefA.Datatype(), sampleA.Value, sampleB.Value)
    return less
}

func (dq *CassDeviceQuery)DeviceList() ([]datalayer.Device, error) {
    devices := []datalayer.Device{}
    var deviceId gocql.UUID
    var accessLevel int

    // Fetch all devices (TODO: inefficient!)
    query := dq.account.conn.session.Query(`
            SELECT device_id, access_level FROM device_permissions 
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

    // Sort
    if dq.sortOrder != nil {
        data := sortData{
            devices: devices,
            sortOrder: dq.sortOrder,
        }
        sort.Sort(data)
    }

    // Apply limits
    out := []datalayer.Device{}
    var i int32
    start := dq.limitStart
    if start < 0 {
        start = 0
    }
    count := dq.limitCount
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
    devices, err := dq.DeviceList()
    if err != nil {
        return 0, err
    }
    return int32(len(devices)), nil
}
