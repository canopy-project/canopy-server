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

package sddl

import (
)

// DatatypeEnum is the datatype of a Cloud Variable
type DatatypeEnum int
const (
    DATATYPE_INVALID DatatypeEnum = iota
    DATATYPE_VOID
    DATATYPE_STRING
    DATATYPE_BOOL
    DATATYPE_INT8
    DATATYPE_UINT8
    DATATYPE_INT16
    DATATYPE_UINT16
    DATATYPE_INT32
    DATATYPE_UINT32
    DATATYPE_FLOAT32
    DATATYPE_FLOAT64
    DATATYPE_DATETIME
    DATATYPE_STRUCT
    DATATYPE_ARRAY
)

// DirectionEnum is the "direction" of a Cloud Variable -- that is, who can
// modify a Cloud Variale (device, cloud, or both)
type DirectionEnum int
const (
    DIRECTION_INVALID DirectionEnum = iota
    DIRECTION_INOUT
    DIRECTION_IN
    DIRECTION_OUT
)

// OptionalityEnum is the "optionality" of a Cloud Variable -- whether or not
// the Cloud Variable can be omitted in transactions.
type OptionalityEnum int
const (
    OPTIONALITY_INVALID OptionalityEnum = iota
    OPTIONALITY_OPTIONAL
    OPTIONALITY_REQUIRED
)

// NumericDisplayHintEnum is a metatdata hint telling the application how to
// display the Cloud Variable's numeric value.
type NumericDisplayHintEnum int
const (
    NUMERIC_DISPLAY_HINT_INVALID NumericDisplayHintEnum = iota
    NUMERIC_DISPLAY_HINT_NORMAL
    NUMERIC_DISPLAY_HINT_PERCENTAGE
    NUMERIC_DISPLAY_HINT_SCIENTIFIC
    NUMERIC_DISPLAY_HINT_HEX
)

// SDDL provides an abstracted interface for working with SDDL content
type SDDL interface {
    // Parse an SDDL document, provided as a golang JSON object.
    ParseDocument(docJson map[string]interface{}) (*Document, error)

    // Parse an SDDL document, provided as a string
    ParseDocumentString(docJson string) (*Document, error)

    // Parse an SDDL variable definition.
    // <decl> is the variable declaration: "optional in float32 temperature".
    // <propsJson> is a golang JSON object containing additional properties:
    //      {
    //          "min-value" : -100,
    //          "max-value" : 150,
    //          "units" : "degrees_c",
    //          ...
    //      }
    ParseVarDef(decl string, propsJson map[string]interface{}) (*VarDef, error)

    // Extend a struct by adding new members
    // TODO: make method of VarDef?
    Extend(varDef VarDef, jsn map[string]interface{}) error

    // Create a new empty SDDL document.
    NewEmptyDocument() Document

    // Create a new Cloud Variable definition for an empty "struct".
    NewEmptyStruct() VarDef
}

// VarDef is a Cloud Variable definition.
// A Cloud Variable definition consists of:
//   - A declaration: "optional out float32 temperature"
//   - Additional properties: {min-value: -100.0}
//   - Child members for composite types (like arrays & structs)
type VarDef interface {
    // Get the datatype of this Cloud Variable, ex: DATATYPE_FLOAT32 or
    // DATATYPE_STRUCT
    Datatype() DatatypeEnum

    // Get the full declaration string, ex: "optional out float32 temperature"
    Declaration() string

    // Get full name of this Cloud Variable, ex: "temperature", "gps.longitude"
    Fullname() string

    // Does this Cloud Variable have a numeric datatype?
    IsNumeric() bool

    // Get a golang JSON representation of this Cloud Variable definition.
    Json() map[string]interface{}

    // Internal routine does the actual work of encoding to JSON
    jsonEncode() (map[string]interface{}, error)

    // Get the "max-value" property for this Cloud Variable, cast to a float64.
    // Returns an error if the Cloud Variable does not have a numeric type.
    MaxValue() (float64, error)

    // Get the "min-value" property for this Cloud Variable, cast to a float64.
    // Returns an error if the Cloud Variable does not have a numeric type.
    MinValue() (float64, error)

    // Get name of this Cloud Variable, ex: "temperature", "longitude"
    Name() string

    // Get the "numeric-display-hint" property for this Cloud Variable.
    // Returns an error if the Cloud Variable does not have a numeric type.
    NumericDisplayHint() (NumericDisplayHintEnum, error)

    // Get the "regex" property, used for string input validation.
    // Returns an error if the Cloud Variable does not have a string type
    Regex() (string, error)

    // Get the children of this Cloud Variable if it is a "struct".
    // Returns an error if the Cloud Variable is not DATATYPE_STRUCT
    StructMembers() ([]VarDef, error)

    // Get a string JSON representation of this Cloud Variable definition.
    ToString() (string, error)

    // Get the "units" property for this Cloud Variable.
    // Returns an error if the Cloud Variable does not have a basic type.
    Units() (string, error)
}

// Document is an SDDL document.  It contains zero or more Cloud Variable
// definitions as well as other metadata.
type Document interface{
    // TODO: other qualifiers & properties?
    AddVarDef(name string, datatype DatatypeEnum) (VarDef, error)

    // Get the document's authors
    Authors() []string

    // Get the document's "description" metadata
    Description() string

    // Add new member variables to this document
    Extend(jsn map[string]interface{}) error

    // Get a golang JSON representation of this SDDL document.
    Json() map[string]interface{}

    // Find a member variable by name
    LookupVarDef(varName string) (VarDef, error)

    // Remove a member variable by name
    // Returns true if removed, false if not found
    RemoveVarDef(varName string) (bool, error)

    // Get a string JSON representation of this SDDL document.
    ToString() (string, error)

    // Get the document's Cloud Variable definitions.
    VarDefs() []VarDef
}

// Singleton Sys gives access to parsing routines:
// sddl.Sys.ParseDocument(...)
var Sys SDDLSys
