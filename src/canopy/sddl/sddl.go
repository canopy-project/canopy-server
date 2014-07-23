package sddl

import (
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
    Name() string
    JustName() string
}

type Control struct {
    name string
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
    description string
    properties []Property
    authors []string
    jsonObj map[string]interface{}
}

func ControlTypeStringToEnum(in string) ControlTypeEnum {
    if in == "trigger" {
        return CONTROL_TYPE_TRIGGER
    } else if in == "parameter" {
        return CONTROL_TYPE_PARAMETER
    }
    return CONTROL_TYPE_INVALID
}

func DatatypeStringToEnum(in string) DatatypeEnum {
    if in == "void" {
        return DATATYPE_VOID
    } else if in == "string" {
        return DATATYPE_STRING
    } else if in == "boolean" {
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
    }
    return DATATYPE_INVALID
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
    return nil, nil
}

func (doc *Document) LookupClass(propName string) (Property, error) {
    return nil, nil
}

func (prop *Control) Name() string {
    return prop.name
}

func (prop *Control) JustName() string {
    return strings.Split(prop.Name(), " ")[1]
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

func (prop *Sensor) JustName() string {
    return strings.Split(prop.Name(), " ")[1]
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

func (prop *Class) Name() string {
    return prop.name
}

func (prop *Class) JustName() string {
    return strings.Split(prop.Name(), " ")[1]
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

func (prop *Class) LookupProperty(propName string) (Property, error) {
    return nil, nil
}

func (prop *Class) LookupClass(propName string) (Property, error) {
    return nil, nil
}

func parseControl(name string, json map[string]interface{}) (*Control, error) {
    prop := Control{name: name}
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

func parseSensor(name string, json map[string]interface{}) (*Sensor, error) {
    prop := Sensor{name: name}
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
