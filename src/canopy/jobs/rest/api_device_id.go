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
package rest

import (
    "canopy/canolog"
    "canopy/cloudvar"
    "canopy/datalayer"
    "canopy/sddl"
    "github.com/gocql/gocql"
    "time"
)

func GET__api__device__id(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    // TODO: Check permissions
    //
    // Used for anonymous devices
    /*deviceIdString := info.URLVars["id"]*/

    /*uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        return nil, URLNotFoundError()
    }*/

    /*device, err := info.Conn.LookupDevice(uuid)
    if err != nil {
        // TODO: What errors to return here?
        return nil, InternalServerError("Device lookup failed")
    }*/
    /*out, err := deviceToJsonObj(info.PigeonSys, device)
    if err != nil {
        return nil, InternalServerError("Generating JSON")
    }*/
    //return out, nil
    // TODO: Use new pigeon system

    return nil, nil
}

func POST__api__device__id(info *RestRequestInfo, sideEffect *RestSideEffects) (map[string]interface{}, RestError) {
    var device datalayer.Device
    var err error

    deviceIdString := info.URLVars["id"]

    if deviceIdString == "self" {
        if info.Device == nil {
            return nil, BadInputError("Expected device credentials with /api/device/self").Log()
        }
        device = info.Device
    } else {
        uuid, err := gocql.ParseUUID(deviceIdString)
        if err != nil {
            return nil, URLNotFoundError()
        }

        // TODO: support anonymous device creation

        if info.Account != nil {
            device, err = info.Account.Device(uuid)
            if err != nil {
                // TODO: What errors to return here?
                return nil, InternalServerError("Device lookup failed").Log()
            }
        } else if info.Device != nil {
            if deviceIdString != string(info.Device.IDString()) {
                // TODO: what error to return?
                return nil, InternalServerError("Device mismatch").Log()
            }
            device = info.Device
        } else {
            return nil, NotLoggedInError()
        }
    }

    // Check for SDDL doc.  If it doesn't exist, then create it.
    // TODO: should this only be done if the device is reporting?
    doc := device.SDDLDocument()
    if doc == nil {
        // Create SDDL for the device if it doesn't exist.
        // TODO: should this be automatically done by device.SDDLClass()?
        newDoc := sddl.Sys.NewEmptyDocument()
        err := device.SetSDDLDocument(newDoc)
        if (err != nil) {
            return nil, InternalServerError("Setting new SDDL document").Log()
        }
        doc = newDoc;
    }

    // Parse payload
    for fieldName, value := range info.BodyObj {
        switch fieldName {
        case "__friendly_name":
            friendlyName, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetName(friendlyName);
        case "__location_note":
            locationNote, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetLocationNote(locationNote);
        case "sddl":
            sddlJsonObj, ok := value.(map[string]interface{})
            if !ok {
                return nil, BadInputError("Expected object \"sddl\"")
            }
            err = device.ExtendSDDL(sddlJsonObj)
            if err != nil {
                return nil, BadInputError(err.Error())
            }
        }
    }

    // Handle vars last
    for fieldName, value := range info.BodyObj {
        switch fieldName {
        case "vars":
            varsJsonObj, ok := value.(map[string]interface{})
            if !ok {
                return nil, BadInputError("Expected object \"vars\"")
            }
            for varName, valueJsonObj := range varsJsonObj {
                varDef, err := device.LookupVarDef(varName)
                if err != nil {
                    canolog.Warn("Cloud variable not found: ", varName)
                    /* TODO: Report warning in response*/
                    continue;
                }

                varVal, err := cloudvar.JsonToCloudVarValue(varDef, valueJsonObj)
                if err != nil {
                    canolog.Warn("Cloud variable value parsing problem: ", varName, err)
                    /* TODO: Report warning in response*/
                    continue;
                }
                device.InsertSample(varDef, time.Now(), varVal);
            }
        }
    }

    /*msg := &pigeon.PigeonMessage {
        Data : info.BodyObj,
    }
    canolog.Info("Sending pigeon message", msg);
    err = info.PigeonSys.SendMessage(deviceIdString, msg, time.Duration(100*time.Millisecond))
    if err != nil {
        canolog.Warn("Problem sending WS message! ", err);
        // TODO: Are there certain errors here that shouldn't be ignored?
        //return nil, rest_errors.NewInternalServerError("SendMessage failed")
    }*/
    // TODO: Use new pigeon system for this

    return map[string]interface{} {
        "result" : "ok",
    }, nil
}
