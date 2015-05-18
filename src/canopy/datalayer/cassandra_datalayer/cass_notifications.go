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
    "time"
    "fmt"
)


type CassNotification struct {
    deviceId string
    t time.Time
    isDismissed bool
    msg string
    notifyType int
}

func (note *CassNotification) Datetime() time.Time {
    return note.t;
}

func (note *CassNotification) Dismiss() error {
    return fmt.Errorf("Not implemented")
}

func (note *CassNotification) IsDismissed() bool {
    return note.isDismissed;
}

func (note *CassNotification) Msg() string {
    return note.msg;
}

func (note *CassNotification) NotifyType() int {
    return note.notifyType;
}
