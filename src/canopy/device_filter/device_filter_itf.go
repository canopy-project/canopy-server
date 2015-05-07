/*
 * Copyright 2015 Canopy Services, Inc.
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
package device_filter

import (
    "canopy/datalayer"
    "fmt"
)

type Compiler interface {
    // Compile an expression, such as "temperature > 48.0" into a
    // DeviceExpressions object.
    Compile(expr string) (DeviceFilter, error)
}

type DeviceFilter interface {
    // Does <device> satisfy the filter criteria?
    SatisfiedBy(device datalayer.Device) (bool, error)

    // Filter list of devices into subset that satisfies this filter.
    Whittle(devices []datalayer.Device) ([]datalayer.Device, error)

    // Count the number of devices in a list of devices that satisfy this filter.
    CountMembers(devices []datalayer.Device) (uint32, error)
}

func NewCompiler() Compiler {
    return &DeviceFilterCompiler{}
}

func Compile(expr string) (DeviceFilter, error) {
    return NewCompiler().Compile(expr)
}


func RunTests() error {
    filter, err := Compile("5 == 4")
    if err != nil {
        return err
    }

    sat, err := filter.SatisfiedBy(nil)
    fmt.Println("SAT: ", sat)
    if err != nil {
        return err
    }
    if !sat {
        return fmt.Errorf("Expectected sat=true")
    }
    return nil
}
