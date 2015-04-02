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

package sddl

import (
    "fmt"
)

func DatatypeEnumToString(in DatatypeEnum) (string, error) {
    switch in {
    case DATATYPE_VOID:
        return "void", nil
    case DATATYPE_STRING:
        return "string", nil
    case DATATYPE_BOOL:
        return "bool", nil
    case DATATYPE_INT8:
        return "int8", nil
    case DATATYPE_UINT8:
        return "uint8", nil
    case DATATYPE_INT16:
        return "int16", nil
    case DATATYPE_UINT16:
        return "uint16", nil
    case DATATYPE_INT32:
        return "int32", nil
    case DATATYPE_UINT32:
        return "uint32", nil
    case DATATYPE_FLOAT32:
        return "float32", nil
    case DATATYPE_FLOAT64:
        return "float64", nil
    case DATATYPE_DATETIME:
        return "datetime", nil
    case DATATYPE_STRUCT:
        return "struct", nil
    case DATATYPE_ARRAY:
        return "array", nil
    default:
        return "", fmt.Errorf("Invalid DatatypeEnum value: ", in)
    }
}

func DatatypeStringToEnum(in string) DatatypeEnum {
    if in == "void" {
        return DATATYPE_VOID
    } else if in == "string" {
        return DATATYPE_STRING
    } else if in == "bool" {
        return DATATYPE_BOOL
    } else if in == "int8" {
        return DATATYPE_INT8
    } else if in == "uint8" {
        return DATATYPE_UINT8
    } else if in == "int16" {
        return DATATYPE_INT16
    } else if in == "uint16" {
        return DATATYPE_UINT16
    } else if in == "int32" {
        return DATATYPE_INT32
    } else if in == "uint32" {
        return DATATYPE_UINT32
    } else if in == "float32" {
        return DATATYPE_FLOAT32
    } else if in == "float64" {
        return DATATYPE_FLOAT64
    } else if in == "datetime" {
        return DATATYPE_DATETIME
    } else if in == "struct" {
        return DATATYPE_STRUCT
    } else if in == "array" {
        return DATATYPE_STRUCT
    }
    return DATATYPE_INVALID
}

func NumericDisplayHintEnumToString(in NumericDisplayHintEnum) (string, error) {
    if in == NUMERIC_DISPLAY_HINT_NORMAL {
        return "normal", nil
    } else if in == NUMERIC_DISPLAY_HINT_PERCENTAGE {
        return "percentage", nil
    } else if in == NUMERIC_DISPLAY_HINT_SCIENTIFIC {
        return "scientific", nil
    } else if in == NUMERIC_DISPLAY_HINT_HEX {
        return "hex", nil
    }
    return "", fmt.Errorf("Invalid NumericDisplayHintEnum value", in)
}

func NumericDisplayHintStringToEnum(in string) NumericDisplayHintEnum {
    if in == "normal" {
        return NUMERIC_DISPLAY_HINT_NORMAL
    } else if in == "percentage" {
        return NUMERIC_DISPLAY_HINT_PERCENTAGE
    } else if in == "scientific" {
        return NUMERIC_DISPLAY_HINT_SCIENTIFIC
    } else if in == "hex" {
        return NUMERIC_DISPLAY_HINT_HEX
    }
    return NUMERIC_DISPLAY_HINT_INVALID
}
