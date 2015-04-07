// Copright 2014-2015 Canopy Services, Inc.
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

package cloudvar

import (
    "canopy/sddl"
    "time"
    "fmt"
)

// CloudVarValue represents the value of a Cloud Variable
//                                                                              
// The dynamic type is determined by the datatype of the Cloud Variable
//                                                                              
//  SDDL DATATYPE                           CloudVarValue GOLANG TYPE            
//  -----------------------------------------------------------------           
//  sddl.DATATYPE_VOID                      interface{}                         
//  sddl.DATATYPE_STRING                    string
//  sddl.DATATYPE_BOOL                      bool
//  sddl.DATATYPE_INT8                      int8
//  sddl.DATATYPE_UINT8                     uint8
//  sddl.DATATYPE_INT16                     int16
//  sddl.DATATYPE_UINT16                    uint16
//  sddl.DATATYPE_INT32                     int32
//  sddl.DATATYPE_UINT32                    uint32
//  sddl.DATATYPE_INT32                     int32
//  sddl.DATATYPE_FLOAT32                   float32
//  sddl.DATATYPE_FLOAT64                   float64
//  sddl.DATATYPE_DATETIME                  time.Time

type CloudVarValue interface {}

type CloudVarSample struct {
    Timestamp time.Time
    Value CloudVarValue
}

type CloudVar struct {
    varDef sddl.VarDef
    value CloudVarValue
}

func Less(datatype sddl.DatatypeEnum, value0, value1 CloudVarValue) (bool, error) {
    switch datatype {
    case sddl.DATATYPE_VOID:
        return false, nil
    case sddl.DATATYPE_STRING:
        v0, ok := value0.(string)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects string value for v0")
        }
        v1, ok := value1.(string)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects string value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_BOOL:
        v0, ok := value0.(bool)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects bool value for v0")
        }
        v1, ok := value1.(bool)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects bool value for v1")
        }
        return (!v0 && v1), nil
    case sddl.DATATYPE_INT8:
        v0, ok := value0.(int8)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects int8 value for v0")
        }
        v1, ok := value1.(int8)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects int8 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_UINT8:
        v0, ok := value0.(uint8)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects uint8 value for v0")
        }
        v1, ok := value1.(uint8)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects uint8 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_INT16:
        v0, ok := value0.(int16)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects int16 value for v0")
        }
        v1, ok := value1.(int16)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects int16 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_UINT16:
        v0, ok := value0.(uint16)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects uint16 value for v0")
        }
        v1, ok := value1.(uint16)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects uint16 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_INT32:
        v0, ok := value0.(int32)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects int32 value for v0")
        }
        v1, ok := value1.(int32)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects int32 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_UINT32:
        v0, ok := value0.(uint32)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects uint32 value for v0")
        }
        v1, ok := value1.(uint32)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects uint32 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_FLOAT32:
        v0, ok := value0.(float32)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects float32 value for v0")
        }
        v1, ok := value1.(float32)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects float32 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_FLOAT64:
        v0, ok := value0.(float64)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects float64 value for v0")
        }
        v1, ok := value1.(float64)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects float64 value for v1")
        }
        return (v0 < v1), nil
    case sddl.DATATYPE_DATETIME:
        v0, ok := value0.(time.Time)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects time.Time value for v0")
        }
        v1, ok := value1.(time.Time)
        if !ok {
            return false, fmt.Errorf("cloudvar.Less expects time.Time value for v1")
        }
        return v0.Before(v1), nil
    default:
        return false, fmt.Errorf("cloudvar.Less unsupported datatype ", datatype)
    }
}

func JsonToCloudVarValue(varDef sddl.VarDef, value interface{}) (interface{}, error) {
    switch varDef.Datatype() {
    case sddl.DATATYPE_VOID:
        return nil, nil
    case sddl.DATATYPE_STRING:
        v, ok := value.(string)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects string value for %s", varDef.Name())
        }
        return v, nil
    case sddl.DATATYPE_BOOL:
        v, ok := value.(bool)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects bool value for %s", varDef.Name())
        }
        return v, nil
    case sddl.DATATYPE_INT8:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return int8(v), nil
    case sddl.DATATYPE_UINT8:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return uint8(v), nil
    case sddl.DATATYPE_INT16:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return int16(v), nil
    case sddl.DATATYPE_UINT16:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return uint16(v), nil
    case sddl.DATATYPE_INT32:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return int32(v), nil
    case sddl.DATATYPE_UINT32:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return uint32(v), nil
    case sddl.DATATYPE_FLOAT32:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return float32(v), nil
    case sddl.DATATYPE_FLOAT64:
        v, ok := value.(float64)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects number value for %s", varDef.Name())
        }
        return v, nil
    case sddl.DATATYPE_DATETIME:
        v, ok := value.(string)
        if !ok {
            return nil, fmt.Errorf("JsonToCloudVarValue expects string value for %s", varDef.Name())
        }
        tval, err := time.Parse(time.RFC3339, v)
        if err != nil {
            return nil, fmt.Errorf("JsonToCloudVarValue expects RFC3339 formatted time value for %s", varDef.Name())
        }
        return tval, nil
    default:
        return nil, fmt.Errorf("InsertSample unsupported datatype ", varDef.Datatype())
    }
}
