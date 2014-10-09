/*
 * Copyright 2014 Gregory Prisament
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
package sddl

import (
    "canopy/canolog"
    "errors"
    "fmt"
    "encoding/json"
    "strings"
)

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
)

type ControlTypeEnum int
const (
    CONTROL_TYPE_INVALID ControlTypeEnum = iota
    CONTROL_TYPE_PARAMETER
    CONTROL_TYPE_TRIGGER
)

type NumericDisplayHintEnum int
const (
    NUMERIC_DISPLAY_HINT_INVALID NumericDisplayHintEnum = iota
    NUMERIC_DISPLAY_HINT_NORMAL
    NUMERIC_DISPLAY_HINT_PERCENTAGE
    NUMERIC_DISPLAY_HINT_SCIENTIFIC
    NUMERIC_DISPLAY_HINT_HEX
)
type Document struct {
    properties []Property
    authors []string
    description string
}

type Property interface {
    Declaration() string
    Name() string
}

type Control struct {
    name string
    decl string
    description string
    datatype DatatypeEnum
    controlType ControlTypeEnum
    maxValue float64
    minValue float64
    numericDisplayHint NumericDisplayHintEnum
    regex string
    units string
}

type Sensor struct {
    name string
    decl string
    description string
    datatype DatatypeEnum
    maxValue float64
    minValue float64
    numericDisplayHint NumericDisplayHintEnum
    regex string
    units string
}

type Class struct {
    name string
    decl string
    description string
    properties []Property
    authors []string
    jsonObj map[string]interface{}
}

func ControlTypeEnumToString(in ControlTypeEnum) (string, error) {
    if in == CONTROL_TYPE_TRIGGER {
        return "trigger", nil
    } else if in == CONTROL_TYPE_PARAMETER {
        return "parameter", nil
    }
    return "", fmt.Errorf("Invalid ControlTypeEnum value: ", in)
}

func ControlTypeStringToEnum(in string) ControlTypeEnum {
    if in == "trigger" {
        return CONTROL_TYPE_TRIGGER
    } else if in == "parameter" {
        return CONTROL_TYPE_PARAMETER
    }
    return CONTROL_TYPE_INVALID
}

func DatatypeEnumToString(in DatatypeEnum) (string, error) {
    if in == DATATYPE_VOID {
        return "void", nil
    } else if in == DATATYPE_STRING {
        return "string", nil
    } else if in == DATATYPE_BOOL {
        return "bool", nil
    } else if in == DATATYPE_INT8 {
        return "int8", nil
    } else if in == DATATYPE_UINT8 {
        return "uint8", nil
    } else if in == DATATYPE_INT16 {
        return "int16", nil
    } else if in == DATATYPE_UINT16 {
        return "uint16", nil
    } else if in == DATATYPE_INT32 {
        return "int32", nil
    } else if in == DATATYPE_UINT32 {
        return "uint32", nil
    } else if in == DATATYPE_FLOAT32 {
        return "float32", nil
    } else if in == DATATYPE_FLOAT64 {
        return "float64", nil
    } else if in == DATATYPE_DATETIME {
        return "datetime", nil
    } else {
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
func (doc *Document) Authors() []string {
    return doc.authors
}

func (doc *Document) Properties() []Property {
    return doc.properties
}

func (doc *Document) LookupProperty(propName string) (Property, error) {
    return nil, fmt.Errorf("Not implement");
}

func (doc *Document) LookupClass(propName string) (Property, error) {
    return nil, fmt.Errorf("Not implement");
}

func (prop *Control) Name() string {
    return prop.name
}

func (prop *Control) Declaration() string {
    return prop.decl
}

func (prop *Control) ControlType() ControlTypeEnum {
    return prop.controlType
}

func (prop *Control) Datatype() DatatypeEnum {
    return prop.datatype
}

func (prop *Control) MaxValue() float64 {
    return prop.maxValue
}

func (prop *Control) MinValue() float64 {
    return prop.minValue
}

func (prop *Control) NumericDisplayHint() NumericDisplayHintEnum {
    return prop.numericDisplayHint
}

func (prop *Control) Units() string {
    return prop.units
}

func (prop *Control) Regex() string {
    return prop.units
}

func (prop *Sensor) Name() string {
    return prop.name
}

func (prop *Sensor) Declaration() string {
    return prop.decl
}

func (prop *Sensor) Datatype() DatatypeEnum {
    return prop.datatype
}

func (prop *Sensor) MaxValue() float64 {
    return prop.maxValue
}

func (prop *Sensor) MinValue() float64 {
    return prop.minValue
}

func (prop *Sensor) NumericDisplayHint() NumericDisplayHintEnum{
    return prop.numericDisplayHint
}

func (prop *Sensor) Units() string {
    return prop.units
}

func (prop *Sensor) Regex() string {
    return prop.units
}

func (prop *Class) AddSensorProperty(propName string, datatype DatatypeEnum) (Property, error) {
    // TODO: What if sensor already exists?

    sensor := Sensor{
        name: propName,
        datatype: datatype,
        decl: "sensor " + propName, // TODO: what does this need to be?
        numericDisplayHint: NUMERIC_DISPLAY_HINT_NORMAL,
    }

    prop.properties = append(prop.properties, &sensor)

    // Re-generate JSON
    jsonObj, err := JsonEncodeClass(prop)
    if err != nil {
        return nil, err
    }
    prop.jsonObj = jsonObj

    return &sensor, nil
}

func (prop *Class) Name() string {
    return prop.name
}

func (prop *Class) Declaration() string {
    return prop.decl
}

func (prop *Class) Json() map[string]interface{} {
    return prop.jsonObj
}

func (prop *Class) ToString() (string, error) {
    out, err := json.Marshal(prop.Json())
    return string(out), err
}

func (prop *Class) Authors() []string {
    return prop.authors
}

func (prop *Class) Description() string {
    return prop.description
}

func (prop *Class) Properties() []Property {
    return prop.properties
}

func (prop *Class) LookupSensor(sensorName string) (*Sensor, error) {
    /* TODO: improve implementation */
    for _, child := range prop.properties {
        sensor, ok := child.(*Sensor)
        if (ok && sensor.Name() == sensorName) {
            return sensor, nil
        }
    }
    return nil, fmt.Errorf("Sensor %s not found in class %s", sensorName, prop.Name())
}

func (prop *Class) LookupProperty(propName string) (Property, error) {
    /* TODO: improve implementation */
    for _, child := range prop.properties {
        if (child.Name() == propName) {
            return child, nil
        }
    }
    return nil, fmt.Errorf("Property %s not found in class %s", propName, prop.Name())
}

func (prop *Class) LookupPropertyOrNil(propName string) (Property) {
    /* TODO: improve implementation */
    /* TODO: We could combine with LookupProperty.. just don't return error
     * when property not found. */
    for _, child := range prop.properties {
        if (child.Name() == propName) {
            return child
        }
    }
    return nil
}

func (prop *Class) LookupClass(propName string) (Property, error) {
    return nil, fmt.Errorf("Not implement");
}

func parseControl(decl string, json map[string]interface{}) (*Control, error) {
    splitDecl := strings.Split(decl, " ");
    if !(len(splitDecl) == 2 && splitDecl[0] == "control") {
        return nil, errors.New("Expected declaration of form: \"control <NAME>\"")
    }
    prop := Control{
        decl: decl, 
        name: splitDecl[1],
        controlType: CONTROL_TYPE_PARAMETER,
        numericDisplayHint: NUMERIC_DISPLAY_HINT_NORMAL,
        // TODO: remaining defaults
    }
    for k, v := range json {
        var ok bool
        if k == "control-type" {
            str, ok := v.(string)
            if !ok {
                return nil, errors.New("Expected string for control-type")
            }
            prop.controlType = ControlTypeStringToEnum(str)
            if prop.controlType == CONTROL_TYPE_INVALID {
                return nil, fmt.Errorf("Invalid control type: ", str)
            }
        } else if k == "datatype" {
            str, ok := v.(string)
            if !ok {
                return nil, errors.New("Expected string for datatype")
            }
            prop.datatype = DatatypeStringToEnum(str)
            if prop.datatype == DATATYPE_INVALID {
                return nil, fmt.Errorf("Invalid datatype type: ", str)
            }
        } else if k == "description" {
            prop.description, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for description")
            }
        } else if k == "max-value" {
            prop.maxValue, ok = v.(float64)
            if !ok {
                return nil, errors.New("Expected number for max-value")
            }
        } else if k == "min-value" {
            prop.minValue, ok = v.(float64)
            if !ok {
                return nil, errors.New("Expected number for min-value")
            }
        } else if k == "numeric-display-hint" {
            hintString, ok := v.(string)
            if !ok {
                return nil, errors.New("Expected string for numeric-display-hint")
            }
            prop.numericDisplayHint = NumericDisplayHintStringToEnum(hintString)
            if prop.numericDisplayHint == NUMERIC_DISPLAY_HINT_INVALID {
                return nil, fmt.Errorf("Invalid numeric display hint: ", hintString)
            }
        } else if k == "regex" {
            prop.regex, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for regex")
            }
        } else if k == "units" {
            prop.units, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for units")
            }
        }
    }
    return &prop, nil
}

func parseSensor(decl string, json map[string]interface{}) (*Sensor, error) {
    splitDecl := strings.Split(decl, " ");
    if !(len(splitDecl) == 2 && splitDecl[0] == "sensor") {
        return nil, errors.New("Expected declaration of form: \"sensor <NAME>\"")
    }
    prop := Sensor{decl: decl, name: splitDecl[1]}
    for k, v := range json {
        var ok bool
        if k == "datatype" {
            str, ok := v.(string)
            if !ok {
                return nil, errors.New("Expected string for datatype")
            }
            prop.datatype = DatatypeStringToEnum(str)
            if prop.datatype == DATATYPE_INVALID {
                return nil, fmt.Errorf("Invalid datatype type: ", str)
            }
        } else if k == "description" {
            prop.description, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for description")
            }
        } else if k == "max-value" {
            prop.maxValue, ok = v.(float64)
            if !ok {
                return nil, errors.New("Expected number for max-value")
            }
        } else if k == "min-value" {
            prop.minValue, ok = v.(float64)
            if !ok {
                return nil, errors.New("Expected number for min-value")
            }
        } else if k == "numeric-display-hint" {
            hintString, ok := v.(string)
            if !ok {
                return nil, errors.New("Expected string for numeric-display-hint")
            }
            prop.numericDisplayHint = NumericDisplayHintStringToEnum(hintString)
            if prop.numericDisplayHint == NUMERIC_DISPLAY_HINT_INVALID {
                return nil, fmt.Errorf("Invalid numeric display hint: ", hintString)
            }
        } else if k == "regex" {
            prop.regex, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for regex")
            }
        } else if k == "units" {
            prop.units, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for units")
            }
        }
    }
    return &prop, nil
}

func ParseClassString(name string, jsonString string) (*Class, error) {
    var data map[string]interface{}

    err := json.Unmarshal([]byte(jsonString), &data)
    if err != nil {
        return nil, err
    }

    return ParseClass(name, data)
}

func ParseClass(name string, jsn map[string]interface{}) (*Class, error) {
    class := Class{
        name: name,
        decl: name,
        properties: []Property{}, 
        authors: []string{},
    }
    class.jsonObj = jsn
    for k, v := range jsn {
        var ok bool
        if strings.HasPrefix(k, "control ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.New("Expected object for control definition")
            }
            control, err := parseControl(k, vObj)
            if err != nil {
                return nil, err
            }
            class.properties = append(class.properties, control)
        } else if strings.HasPrefix(k, "sensor ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.New("Expected object for sensor definition")
            }
            sensor, err := parseSensor(k, vObj)
            if err != nil {
                return nil, err
            }
            class.properties = append(class.properties, sensor)
        } else if strings.HasPrefix(k, "class ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.New("Expected object for class definition")
            }
            childClass, err := ParseClass(k, vObj)
            if err != nil {
                return nil, err
            }
            class.properties = append(class.properties, childClass)
        } else if k == "authors" {
            authorsList, ok := v.([]interface{})
            /* TODO: finish */
            if !ok {
                return nil, errors.New("Expected list for authors")
            }
            for _, authorItf := range authorsList {
                authorString, ok := authorItf.(string)
                if !ok {
                    return nil, errors.New("Expect string for author")
                }
                class.authors = append(class.authors, authorString)
            }
        } else if k == "description" {
            class.description, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for description")
            }
        }
    }

    return &class, nil
}

func ExtendClass(class *Class, jsn map[string]interface{}) error {
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
        } else if strings.HasPrefix(k, "sensor ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return errors.New("Expected object for sensor definition")
            }
            sensor, err := parseSensor(k, vObj)
            if err != nil {
                return err
            }
            class.properties = append(class.properties, sensor)
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
}

func ParseDocument(jsn map[string]interface{}) (*Document, error) {
    doc := Document{properties: []Property{}, authors: []string{}}
    for k, v := range jsn {
        var ok bool
        if strings.HasPrefix(k, "control ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.New("Expected object for control definition")
            }
            control, err := parseControl(k, vObj)
            if err != nil {
                return nil, err
            }
            doc.properties = append(doc.properties, control)
        } else if strings.HasPrefix(k, "sensor ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.New("Expected object for sensor definition")
            }
            sensor, err := parseSensor(k, vObj)
            if err != nil {
                return nil, err
            }
            doc.properties = append(doc.properties, sensor)
        } else if strings.HasPrefix(k, "class ") {
            vObj, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.New("Expected object for doc definition")
            }
            class, err := ParseClass(k, vObj)
            if err != nil {
                return nil, err
            }
            doc.properties = append(doc.properties, class)
        } else if k == "authors" {
            authorsList, ok := v.([]interface{})
            /* TODO: finish */
            if !ok {
                return nil, errors.New("Expected list for authors")
            }
            for _, authorItf := range authorsList {
                authorString, ok := authorItf.(string)
                if !ok {
                    return nil, errors.New("Expect string for author")
                }
                doc.authors = append(doc.authors, authorString)
            }
        } else if k == "description" {
            doc.description, ok = v.(string)
            if !ok {
                return nil, errors.New("Expected string for description")
            }
        }
    }

    return &doc, nil
}

func NewEmptyClass() (*Class) {
    var class Class;
    return &class;
}

func jsonEncodeProperty(prop Property) (map[string]interface{}, error) {
    jsn := map[string]interface{}{}

    switch p := prop.(type) {
    case *Control:
        jsn["description"] = p.description

        datatype, err := DatatypeEnumToString(p.datatype)
        if err != nil {
            return nil, err
        }
        jsn["datatype"] = datatype

        controlType, err := ControlTypeEnumToString(p.controlType)
        if err != nil {
            return nil, err
        }
        jsn["control-type"] = controlType

        // TODO: Don't always have these?
        jsn["max-value"] = p.maxValue
        jsn["min-value"] = p.minValue

        numericDisplayHint, err := NumericDisplayHintEnumToString(p.numericDisplayHint)
        if err != nil {
            return nil, err
        }
        jsn["numeric-display-hint"] = numericDisplayHint

        jsn["regex"] = p.regex
        jsn["units"] = p.units
    case *Sensor:
        jsn["description"] = p.description

        datatype, err := DatatypeEnumToString(p.datatype)
        if err != nil {
            return nil, err
        }
        jsn["datatype"] = datatype

        // TODO: Don't always have these?
        jsn["max-value"] = p.maxValue
        jsn["min-value"] = p.minValue

        numericDisplayHint, err := NumericDisplayHintEnumToString(p.numericDisplayHint)
        if err != nil {
            return nil, err
        }
        jsn["numeric-display-hint"] = numericDisplayHint

        jsn["regex"] = p.regex
        jsn["units"] = p.units

        // TODO: recursive case for encoding sub-classes
    default:
        return nil, fmt.Errorf("jsonEncodeProperty expects control or sensor property")
    }

    return jsn, nil
}

func JsonEncodeClass(cls *Class) (map[string]interface{}, error) {
    jsn := map[string]interface{}{}

    for _, prop := range cls.properties {
        val, err := jsonEncodeProperty(prop)
        if err != nil {
            return nil, err
        }
        jsn[prop.Declaration()] = val
    }
    // TODO: authors
    // TODO: description
    return jsn, nil
}

func JsonStringEncodeClass(cls *Class) (string, error) {
    jsn, err := JsonEncodeClass(cls)
    if err != nil {
        return "", err
    }

    bytes, err := json.Marshal(jsn)
    if err != nil {
        return "", err
    }

    return string(bytes), nil
}
