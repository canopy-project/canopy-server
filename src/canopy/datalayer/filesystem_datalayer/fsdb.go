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

//
// Filesystem database:
//
// <path>/data/accounts/<username>
// <path>/data/devices/<id>

type FsdbDatalayer struct {
    cfg config.Config
    path string // i.e. "/var/canopy/db"
}
func (dl *FsdbDatalayer) Connect() datalayer.Connection, error {
    return &FsdbConnection{}, nil
}

func (dl *FsdbDatalayer) datapath() string {
    // We store everything in the (<path> + "/data") directory.  This is mostly
    // for damage control if <path> is misconfigured.  This way, we'll never,
    // for example "rm -rf /" wiping the whole file system.
    return path + "/data"
}

func (dl *FsdbDatalayer) EraseDb() error {
    os.RemoveAll(dl.datapath())
}

func (dl *FsdbDatalayer) PrepDb() error {
    return os.MkdirAll(dl.datapath(), 0777)
}

func (dl *FsdbDatalayer) MigrateDb() error {
    return fmt.Errorf("Not implemented")
}

func NewDatalayer(cfg config.Config) datalayer.Datalayer {
    return &FsdbDatalayer{cfg: cfg, path: "/var/canopy/db"}
}
