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

package cloudvar

import (
    "canopy/sddl"
    "time"
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
