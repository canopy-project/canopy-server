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
package endpoints

//
// The device POSTs to this endpoint when a firmware routine such as
// "canopy_easy_post_sample" is called.
//
//  POST /di/device/<UUID>
//  {
//      <sample_name> : <value>,
//      ...
//  }
//
//  Effects:
//
//  - If <UUID> does not correspond to a device, an "anonymous" device is
//    created.
//  - If <UUID> corresponds to a non-anonymous device, the request will fail
//    unless property authenticated.
//
//  - Any posted samples that do not correspond to an existing SDDL property,
//    that SDDL property is created.
//
//
//  POST /di/device/<UUID>
//  {
//      "__sddl_update" : {
//          "control foo" : {
//              "datatype" : "float32",
//          }
//      }
//  }
//  - Adds control if it doesn't already exist.
//
//

import (
    "canopy/datalayer"
    "canopy/datalayer/cassandra_datalayer"
    "canopy/canolog"
    "canopy/cloudvar"
    "canopy/config"
    "canopy/sddl"
    "encoding/json"
    "fmt"
    "github.com/gocql/gocql"
    "github.com/gorilla/mux"
    "net/http"
    "strings"
    "time"
)

// This is a bit of a hack to communicate the server's configuration to this
// endpoint.  Is this endpoint even still used?
var globalConfig config.Config
func SetGlobalConfig(cfg config.Config) {
    globalConfig = cfg
}
func GetGlobalConfig() config.Config{
    return globalConfig
}

func POST_di__device__id(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    deviceIdString := vars["id"]
    canolog.Info("/di/device/", deviceIdString, " requested.")

    // Parse input as json
    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    // Connect to database
    dl := cassandra_datalayer.NewDatalayer(GetGlobalConfig())
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    // Parse UUID
    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device UUID expected\"}");
        return
    }

    // Does device exist?  If not, create an anonymous device.
    device, err := conn.LookupOrCreateDevice(uuid, datalayer.ReadWriteAccess)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Error reading database\"}");
        return
    }
    doc := device.SDDLDocument()
    canolog.Info("Checking for SDDL")
    if doc == nil {
        canolog.Info("SDDL Not found")
        // Create SDDL for the device if it doesn't exist.
        // TODO: should this be automatically done by device.SDDLClass()?
        newDoc := sddl.Sys.NewEmptyDocument()
        err = device.SetSDDLDocument(newDoc)
        canolog.Info("Setting SDDL")
        if (err != nil) {
            canolog.Info("Oops -error")
            w.WriteHeader(http.StatusInternalServerError);
            fmt.Fprintf(w, "{\"error\" : \"Error creating SDDL for new device\"}");
            return
        }
        doc = newDoc;
    }

    // For each reported property, create SDDL property if necessary
    for varName, value := range data {
        if strings.HasPrefix(varName, "__") {
            continue;
        }
        varDef, err := doc.LookupVarDef(varName)
        // TODO: an error doesn't necessarily mean prop should be created?
        canolog.Info("Looking up property ", varName)
        if (varDef == nil) {
            // Property doesn't exist.  Add it.
            canolog.Info("Not found.  Add property ", varName)
            // TODO: What datatype?
            // TODO: What other parameters?
            varDef, err = doc.AddVarDef(varName, sddl.DATATYPE_FLOAT32)
            if err != nil {
                canolog.Info("Oops error", err)
                w.WriteHeader(http.StatusInternalServerError);
                fmt.Fprintf(w, "{\"error\" : \"Error adding SDDL property\"}");
                return
            }

            // save modified SDDL 
            canolog.Info("SetSDDLDocument ", doc)
            err = device.SetSDDLDocument(doc)
            if err != nil {
                canolog.Info("Oops error", err)
                w.WriteHeader(http.StatusInternalServerError);
                fmt.Fprintf(w, "{\"error\" : \"Error Updating SDDL\"}");
                return
            }
        }

        // Store property value.
        // Convert value datatype
        varVal, err := cloudvar.JsonToCloudVarValue(varDef, value)
        if err != nil {
                canolog.Info("Error converting JSON to property value: ", value)
                w.WriteHeader(http.StatusInternalServerError);
                fmt.Fprintf(w, "{\"error\" : \"Error converting JSON to property value\"}");
                return
        }
        canolog.Info("InsertStample")
        err = device.InsertSample(varDef, time.Now(), varVal)
        if (err != nil) {
            canolog.Warn("Error inserting sample ", varName, ": ", err)
        }
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\"}");
}
