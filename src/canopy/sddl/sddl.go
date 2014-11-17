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
    "canopy/canolog"
    "errors"
    "fmt"
    "encoding/json"
    "strings"
)

type SDDLDocument struct {
    vars []VarDef
    authors []string
    description string
    jsonObj map[string]interface{}
}

type SDDLVarDef struct {
    name string
    decl string
    description string
    datatype DatatypeEnum
    optionality OptionalityEnum
    direction DirectionEnum
    maxValue float64
    minValue float64
    numericDisplayHint NumericDisplayHintEnum
    regex string
    units string
    structVars []VarDef
    arraySize int
    arrayElement *SDDLVarDef
    jsonObj map[string]interface{}
}

type SDDLSys struct {}

// Helper routine for parsing defininition keywords
func keyTokenFromString(s string) (string, int, error) {
    switch s {
    case "void":
        return "datatype", int(DATATYPE_VOID), nil
    case "bool":
        return "datatype", int(DATATYPE_BOOL), nil
    case "int8":
        return "datatype", int(DATATYPE_INT8), nil
    case "uint8":
        return "datatype", int(DATATYPE_UINT8), nil
    case "int16":
        return "datatype", int(DATATYPE_INT16), nil
    case "uint16":
        return "datatype", int(DATATYPE_UINT16), nil
    case "int32":
        return "datatype", int(DATATYPE_INT32), nil
    case "uint32":
        return "datatype", int(DATATYPE_UINT32), nil
    case "float32":
        return "datatype", int(DATATYPE_FLOAT32), nil
    case "float64":
        return "datatype", int(DATATYPE_FLOAT64), nil
    case "string":
        return "datatype", int(DATATYPE_STRING), nil
    case "datetime":
        return "datatype", int(DATATYPE_DATETIME), nil

    case "bidirectional":
        return "direction", int(DIRECTION_BIDIRECTIONAL), nil
    case "inbound":
        return "direction", int(DIRECTION_INBOUND), nil
    case "outbound":
        return "direction", int(DIRECTION_OUTBOUND), nil

    case "optional":
        return "optionality", int(OPTIONALITY_OPTIONAL), nil
    case "required":
        return "optionality", int(OPTIONALITY_REQUIRED), nil
    }
    return "unknown", 0, nil 
}

// Helper routine for parsing defininition strings
func parseVarKey(key string) (OptionalityEnum, DirectionEnum, DatatypeEnum, string, error) {
    optionality := OPTIONALITY_INVALID;
    direction := DIRECTION_INVALID;
    datatype := DATATYPE_INVALID;
    name := ""

    parts := strings.Split(key, " ")

    for _, part := range parts {
        tokenType, tokenVal, err := keyTokenFromString(part)
        if err != nil {
            return 0, 0, 0, "", err
        }

        switch tokenType {
            case "datatype" :
                if datatype != DATATYPE_INVALID {
                    return 0, 0, 0, "", fmt.Errorf("Datatype already specified")
                }
                datatype = DatatypeEnum(tokenVal)
            case "direction" :
                if direction != DIRECTION_INVALID {
                    return 0, 0, 0, "", fmt.Errorf("Direction already specified")
                }
                direction = DirectionEnum(tokenVal)
            case "optionality" :
                if optionality != OPTIONALITY_INVALID {
                    return 0, 0, 0, "", fmt.Errorf("Optionality already specified")
                }
                optionality = OptionalityEnum(tokenVal)
            default:
                if datatype == DATATYPE_INVALID {
                    return 0, 0, 0, "", fmt.Errorf("Datatype or qualifier expected: %s", part)
                }
                if name != "" {
                    return 0, 0, 0, "", fmt.Errorf("Variable name already specified")
                }
                name = part
        }
    }
    if name == "" {
            return 0, 0, 0, "", fmt.Errorf("Variable name expected")
    }

    return optionality, direction, datatype, name, nil
}

func ParseVar(decl string, defJson map[string]interface{}) (VarDef, error) {
    optionality, direction, datatype, name, err := parseVarKey(decl)
    if err != nil {
        return nil, err
    }

    varDef := SDDLVarDef{
        decl: decl, 
        name: name,
        optionality: optionality,
        direction: direction,
        datatype: datatype,
        numericDisplayHint: NUMERIC_DISPLAY_HINT_NORMAL,
        // TODO: remaining defaults
    }
    for k, v := range defJson {
        var ok bool
        if k == "description" {
            varDef.description, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for description")
            }
        } else if k == "max-value" {
            varDef.maxValue, ok = v.(float64)
            if !ok {
                return nil, errors.New("Expected number for max-value")
            }
        } else if k == "min-value" {
            varDef.minValue, ok = v.(float64)
            if !ok {
                return nil, errors.New("Expected number for min-value")
            }
        } else if k == "numeric-display-hint" {
            hintString, ok := v.(string)
            if !ok {
                return nil, errors.New("Expected string for numeric-display-hint")
            }
            varDef.numericDisplayHint = NumericDisplayHintStringToEnum(hintString)
            if varDef.numericDisplayHint == NUMERIC_DISPLAY_HINT_INVALID {
                return nil, fmt.Errorf("Invalid numeric display hint: ", hintString)
            }
        } else if k == "regex" {
            varDef.regex, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for regex")
            }
        } else if k == "units" {
            varDef.units, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for units")
            }
        }
    }
    return &varDef, nil
}

func (sys *SDDLSys) ParseDocument(jsn map[string]interface{}) (Document, error) {
    doc := SDDLDocument{
        jsonObj: jsn, 
        vars: []VarDef{}, 
        authors: []string{},
    }

    err := doc.Extend(jsn)
    if err != nil {
        return nil, err
    }
    return &doc, nil
}

func (sys *SDDLSys) ParseDocumentString(doc string) (Document, error) {
    var jsn map[string]interface{}
    err := json.Unmarshal([]byte(doc), &jsn)
    if err != nil {
        canolog.Error("Error JSON decoding SDDL docoument: %s %s", doc, err)
        return nil, err
    }
    return sys.ParseDocument(jsn)
}

func (sys *SDDLSys) NewEmptyDocument() (Document) {
    var doc SDDLDocument;
    return &doc;
}

func (sys *SDDLSys) NewEmptyStruct() (VarDef) {
    var varDef SDDLVarDef;
    varDef.datatype = DATATYPE_STRUCT
    // TODO: init other fields
    return &varDef;
}

func (varDef *SDDLVarDef) Datatype() DatatypeEnum {
    return varDef.datatype
}

func (varDef *SDDLVarDef) Declaration() string {
    return varDef.decl
}

func (varDef *SDDLVarDef) Fullname() string {
    return varDef.name // TODO: implement correctly
}

func (varDef *SDDLVarDef) IsNumeric() bool {
    return ((varDef.datatype == DATATYPE_FLOAT32) || (varDef.datatype == DATATYPE_FLOAT64) || (varDef.datatype == DATATYPE_INT8) || (varDef.datatype == DATATYPE_INT16) || (varDef.datatype == DATATYPE_INT32) || (varDef.datatype == DATATYPE_UINT8) || (varDef.datatype == DATATYPE_UINT16) || (varDef.datatype == DATATYPE_UINT32))
}

func (varDef *SDDLVarDef) Json() map[string]interface{} {
    return varDef.jsonObj
}

// Encode Cloud Variable definition as golang JSON object.
// Does the actual work of encoding/marshalling, unlike .Json() which just
// returns cached version.
func (varDef *SDDLVarDef) jsonEncode() (map[string]interface{}, error) {
    jsn := map[string]interface{}{}

    jsn["description"] = varDef.description

    datatype, err := DatatypeEnumToString(varDef.datatype)
    if err != nil {
        return nil, err
    }
    jsn["datatype"] = datatype

    // TODO: Don't always have these?
    if varDef.IsNumeric() {
        jsn["max-value"] = varDef.maxValue
        jsn["min-value"] = varDef.minValue

        numericDisplayHint, err := NumericDisplayHintEnumToString(varDef.numericDisplayHint)
        if err != nil {
            return nil, err
        }
        jsn["numeric-display-hint"] = numericDisplayHint
    }

    if varDef.datatype == DATATYPE_STRING {
        jsn["regex"] = varDef.regex
    }

    jsn["units"] = varDef.units

    return jsn, nil
}


// Encode a Cloud Variable definition as a JSON string.
func (varDef *SDDLVarDef) jsonStringEncode() (string, error) {
    jsn, err := varDef.jsonEncode()
    if err != nil {
        return "", err
    }

    bytes, err := json.Marshal(jsn)
    if err != nil {
        return "", err
    }

    return string(bytes), nil
}


func (varDef *SDDLVarDef) MaxValue() (float64, error)  {
    if !varDef.IsNumeric() {
        return 0, fmt.Errorf("MaxValue() can only be called on a numeric var")
    }
    return varDef.maxValue, nil
}

func (varDef *SDDLVarDef) MinValue() (float64, error)  {
    if !varDef.IsNumeric() {
        return 0, fmt.Errorf("MaxValue() can only be called on a numeric var")
    }
    return varDef.minValue, nil
}

func (varDef *SDDLVarDef) Name() string {
    return varDef.name
}

func (varDef *SDDLVarDef) NumericDisplayHint() (NumericDisplayHintEnum, error) {
    if !varDef.IsNumeric() {
        return 0, fmt.Errorf("NumericDisplayHint() can only be called on a numeric var")
    }
    return varDef.numericDisplayHint, nil
}

func (varDef *SDDLVarDef) Regex() (string, error) {
    if varDef.datatype != DATATYPE_STRING {
        return "", fmt.Errorf("Regex() can only be called on a string var")
    }
    return varDef.regex, nil
}

func (varDef *SDDLVarDef) StructMembers() ([]VarDef, error) {
    if varDef.datatype != DATATYPE_STRUCT {
        return nil, fmt.Errorf("StructMembers() can only be called on a structure")
    }
    return varDef.structVars, nil
}

func (varDef *SDDLVarDef) ToString() (string, error) {
    out, err := json.Marshal(varDef.Json())
    return string(out), err
}

func (varDef *SDDLVarDef) Units() (string, error) {
    return varDef.units, nil
}

/*func ExtendClass(class *Class, jsn map[string]interface{}) error {
    // TODO: combine implementation with ParseClass ?
    for k, v := range jsn {
        canolog.Info("Key:", k)
        if strings.HasPrefix(k, "control ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                canolog.Info("Expected object for control definition")
                return errors.New("Expected object for control definition")
            }
            control, err := parseControl(k, vObj)
            if err != nil {
                canolog.Info("Error: ", err)
                return err
            }
            class.properties = append(class.properties, control)
            canolog.Info("Control added")
        } else if strings.HasPrefix(k, "class ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return errors.New("Expected object for class definition")
            }
            childClass, err := ParseClass(k, vObj)
            if err != nil {
                return err
            }
            class.properties = append(class.properties, childClass)
        }
    }

    // Re-generate JSON
    jsonObj, err := JsonEncodeClass(class)
    if err != nil {
        return err
    }
    class.jsonObj = jsonObj

    return nil
}*/



func (doc *SDDLDocument) AddVarDef(name string, datatype DatatypeEnum) (VarDef, error) {
    // TODO: What if cloud variable already exists?

    datatypeString, err := DatatypeEnumToString(datatype)
    if err != nil {
        return nil, err
    }

    varDef := &SDDLVarDef{
        name: name,
        datatype: datatype,
        decl: datatypeString + " " + name,
        numericDisplayHint: NUMERIC_DISPLAY_HINT_NORMAL,
    }

    doc.vars = append(doc.vars, varDef)

    // Re-generate JSON
    jsonObj, err := varDef.jsonEncode()
    if err != nil {
        return nil, err
    }
    varDef.jsonObj = jsonObj

    return varDef, nil
}

func (doc *SDDLDocument) Authors() []string {
    return doc.authors
}

func (doc *SDDLDocument) Description() string {
    return doc.description
}

func (doc *SDDLDocument) Extend(jsn map[string]interface{}) error {
    var ok bool

    for k, v := range jsn {
        if k == "authors" {
            authorsList, ok := v.([]interface{})
            /* TODO: finish */
            if !ok {
                return errors.New("Expected list for authors")
            }
            for _, authorItf := range authorsList {
                authorString, ok := authorItf.(string)
                if !ok {
                    return errors.New("Expect string for author")
                }
                doc.authors = append(doc.authors, authorString)
            }
        } else if k == "description" {
            doc.description, ok = v.(string)
            if !ok {
                return errors.New("Expected string for description")
            }
        } else {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return errors.New("Expected object for variable metadata")
            }
            varDef, err := ParseVar(k, vObj);
            if (err != nil) {
                return err;
            }

            doc.RemoveVarDef(varDef.Name()) // If var already exists, remove it first
            doc.vars = append(doc.vars, varDef);
        }
    }

    return nil
}

// Encode document as golang JSON object.
// Does the actual work of encoding/marshalling, unlike .Json() which just re
func (doc *SDDLDocument) jsonEncode() (map[string]interface{}, error) {
    jsn := map[string]interface{}{}

    for _, varDef := range doc.vars {
        val, err := varDef.jsonEncode()
        if err != nil {
            return nil, err
        }
        jsn[varDef.Declaration()] = val
    }
    // TODO: authors
    // TODO: description
    return jsn, nil
}

// Encode document as string JSON object.
func (doc *SDDLDocument) jsonStringEncode() (string, error) {
    jsn, err := doc.jsonEncode()
    if err != nil {
        return "", err
    }

    bytes, err := json.Marshal(jsn)
    if err != nil {
        return "", err
    }

    return string(bytes), nil
}

func (doc *SDDLDocument) Json() map[string]interface{} {
    return doc.jsonObj
}

func (doc *SDDLDocument) LookupVarDef(varName string) (VarDef, error) {
    // TODO: improve implementation
    for _, varDef := range doc.vars {
        if (varDef.Name() == varName) {
            return varDef, nil
        }
    }
    return nil, fmt.Errorf("Variable %s not found in document", varName)
}

func (doc *SDDLDocument) RemoveVarDef(varName string) (bool, error) {
    // TODO: improve implementation
    var removed bool
    newVars := []VarDef{}

    for _, varDef := range doc.vars {
        if (varDef.Name() != varName) {
            newVars = append(newVars, varDef);
        } else {
            removed = true
        }
    }

    doc.vars = newVars
    return removed, nil
}

func (doc *SDDLDocument) ToString() (string, error) {
    out, err := json.Marshal(doc.Json())
    return string(out), err
}

func (doc *SDDLDocument) VarDefs() []VarDef{
    return doc.vars
}

