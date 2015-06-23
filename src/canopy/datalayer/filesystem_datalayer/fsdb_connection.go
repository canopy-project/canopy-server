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

import os

func (conn *FsdbConnection) ClearSensorData() error {
    // TODO: implement
}

func (conn *FsdbConnection) Close() {
    // noop
}

func (conn *FsdbConnection) DeleteAccount(username string) {
    return os.Remove(dl.datapath() + "/accounts/" + username)
}

func (conn *FsdbConnection) DeleteDevice(deviceId gocql.UUID) {
    return os.Remove(dl.datapath() + "/devices/" + deviceId.ToString())
}

func (conn *FsdbConnection) LookupAccount(usernameOrEmail string) (Account, error) {
    // TODO: handle email addresses
    file, err := os.Open(dl.datapath()  + "/accounts/" + usernameOrEmail)
    if err != nil {
        return nil, err
    }

    // Read file and parse JSON
    account = &FsdbAccount{conn: conn}
    err = json.NewDecoder(file).Decode(account)
    if err != nil {
        return err
    }

    return account
}

