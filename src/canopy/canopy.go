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
package main

import (
    "errors"
    "fmt"
    "net/http"
    "code.google.com/p/go.net/websocket"
    "github.com/gocql/gocql"
    "github.com/gorilla/sessions"
    "github.com/gorilla/context"
    "github.com/gorilla/mux"
    "canopy/datalayer"
    "canopy/mail"
    "canopy/pigeon"
    "encoding/json"
    "encoding/base64"
    "os"
    "strings"
    "time"
)

func writeNotLoggedInError(w http.ResponseWriter) {
    w.WriteHeader(http.StatusUnauthorized);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "not_logged_in"}`);
}

func writeAccountLookupFailedError(w http.ResponseWriter) {
    w.WriteHeader(http.StatusInternalServerError);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "account_lookup_failed"}`);
}

func writeIncorrectUsernameOrPasswordError(w http.ResponseWriter) {
    w.WriteHeader(http.StatusUnauthorized);
    fmt.Fprintf(w, `{"result" : "error", "error_type" : "incorrect_username_or_password"}`);
}

func writeStandardHeaders(w http.ResponseWriter) {
    w.Header().Set("Connection", "close")
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", gConfAllowOrigin)
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func basicAuthFromRequest(r *http.Request) (username string, password string, err error) {
    h, ok := r.Header["Authorization"]
    if !ok || len(h) == 0 {
        return "", "", errors.New("Authorization header not set")
    }
    parts := strings.SplitN(h[0], " ", 2)
    if len(parts) != 2 {
        return "", "", errors.New("Authentication header malformed")
    }
    if parts[0] != "Basic" {
        return "", "", errors.New("Expected basic authentication")
    }
    encodedVal := parts[1]
    decodedVal, err := base64.StdEncoding.DecodeString(encodedVal)
    if err != nil {
        return "", "", errors.New("Authentication header malformed")
    }
    parts = strings.Split(string(decodedVal), ":")
    if len(parts) != 2 {
        return "", "", errors.New("Authentication header malformed")
    }
    return parts[0], parts[1], nil
}

var store = sessions.NewCookieStore([]byte("my_production_secret"))

func loginHandler(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);

    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    username, ok := data["username"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"string_username_expected\"}")
        return
    }

    password, ok := data["password"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"string_password_expected\"}")
        return
    }

    session, _ := store.Get(r, "canopy-login-session")
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()
    account, err := dl.LookupAccountVerifyPassword(username, password)
    if err == nil {
        session.Values["logged_in_username"] = username
        err := session.Save(r, w)
        if err != nil {
            fmt.Fprintf(w, "{\"error\" : \"saving_session\"}")
            return
        }
        fmt.Fprintf(w, "{\"result\" : \"ok\", \"username\" : \"%s\", \"email\" : \"%s\"}",
            account.Username(),
            account.Email())
        return
    } else {
        writeIncorrectUsernameOrPasswordError(w);
        return
    }
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")
    session.Values["logged_in_username"] = ""
    err := session.Save(r, w)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{ \"error\" : \"could_not_logout\"");
        return;
    }
    fmt.Fprintf(w, "{ \"success\" : true }")
}

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);

    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    username, ok := data["username"].(string)
    if !ok {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"string_username_expected\"}")
        return
    }

    email, ok := data["username"].(string)
    if !ok {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"string_email_expected\"}")
        return
    }

    password, ok := data["password"].(string)
    if !ok {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"string_password_expected\"}")
        return
    }

    password_confirm, ok := data["password_confirm"].(string)
    if !ok {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"string_password_confirm_expected\"}")
        return
    }

    if (password != password_confirm) {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"passwords_dont_match\"}")
        return
    }

    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()

    dl.CreateAccount(username, email, password);
    session, _ := store.Get(r, "canopy-login-session")
    session.Values["logged_in_username"] = username
    err = session.Save(r, w)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"saving_session\"}")
        return
    }
    fmt.Fprintf(w, "{\"success\" : true}")
    return
}

func createDeviceHandler(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);

    username, password, err := basicAuthFromRequest(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        fmt.Fprintf(w, "{\"error\" : \"bad_credentials\"}")
        return
    }

    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()

    acct, err := dl.LookupAccountVerifyPassword(username, password)
    if err != nil {
        if err == datalayer.InvalidPasswordError {
            w.WriteHeader(http.StatusUnauthorized)
            fmt.Fprintf(w, "{\"error\" : \"incorrect_username_or_password\"}")
            return;
        } else {
            w.WriteHeader(http.StatusInternalServerError);
            fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
            return
        }
    }
    
    device, err := dl.CreateDevice("Pending Device");
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"device_creation_failed\"}");
        return
    }

    //err = device.SetAccountAccess(acct, datalayer.ReadWriteShareAccess);
    err = device.SetAccountAccess(acct, 4);
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"could_not_grant_access\"}");
        return
    }

    fmt.Fprintf(w, "{\"success\" : true, \"device_id\" : \"%s\"}", device.GetId().String())
    return
}

func meHandler(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")
    
    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            writeNotLoggedInError(w);
            return
        }
    } else {
        writeNotLoggedInError(w);
        return
    }
    
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()

    account, err := dl.LookupAccount(username_string)
    if err != nil {
        return
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\", \"username\" : \"%s\", \"email\" : \"%s\"}",
        account.Username(),
        account.Email())
    return
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")
    
    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            writeNotLoggedInError(w);
            return
        }
    } else {
        writeNotLoggedInError(w);
        return
    }
    
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()

    account, err := dl.LookupAccount(username_string)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return
    }

    devices, err := account.GetDevices()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"device_lookup_failed\"}");
        return
    }
    out, err := devicesToJson(devices)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"generating_json\"}");
        return
    }
    fmt.Fprintf(w, out);
  
    return 
}

func sensorDataHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deviceIdString := vars["id"]
    sensorName := vars["sensor"]

    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")
    
    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            writeNotLoggedInError(w);
            return
        }
    } else {
        writeNotLoggedInError(w);
        return
    }
    
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()
    account, err := dl.LookupAccount(username_string)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return
    }

    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device UUID expected\"}");
        return
    }

    device, err := account.GetDeviceById(uuid)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Could not find or access device\"}");
        return
    }

    samples, err := device.GetSensorData(sensorName, time.Now(), time.Now())
    if err != nil {
        fmt.Println(err)
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Could not obtain sample data\"}");
        return
    }

    out, err := samplesToJson(samples)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"generating_json\"} : ", err);
        return
    }

    fmt.Fprintf(w, out);
    return 
}

func controlHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deviceIdString := vars["id"]
    //controlName := vars["control"]

    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")
    
    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            writeNotLoggedInError(w);
            return
        }
    } else {
        w.WriteHeader(http.StatusUnauthorized);
        fmt.Fprintf(w, "{\"error\" : \"not_logged_in2\"}");
        return
    }
    
    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()
    account, err := dl.LookupAccount(username_string)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return
    }

    uuid, err := gocql.ParseUUID(deviceIdString)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Device UUID expected\"}");
        return
    }

    device, err := account.GetDeviceById(uuid)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"Could not find or access device\"}");
        return
    }

    /* Parse input as json and just forward it along using pigeon */
    var data map[string]interface{}
    decoder := json.NewDecoder(r.Body)
    err = decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    /* Store control value.  For now, use sensor_data table */
    for sensorName, value := range data {
        /* TODO: Verify that control is, in fact, a control according to SDDL
         * class */
        if (sensorName == "__friendly_name") {
            friendlyName, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetFriendlyName(friendlyName);
        } else if (sensorName == "__location_note") {
            locationNote, ok := value.(string)
            if !ok {
                continue;
            }
            device.SetLocationNote(locationNote);
        } else {
            floatVal, ok := value.(float64)
            if !ok {
                continue;
            }
            /* TODO: fix this! */
            device.InsertSensorSample_float64(sensorName, time.Now(), floatVal);
        }
    }

    msg := &pigeon.PigeonMessage { 
        Data : data,
    }
    err = gPigeon.SendMessage(deviceIdString, msg, time.Duration(100*time.Millisecond))
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"SendMessage failed\"}");
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\"}");
    return 
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
    /*
     *  POST
     *  {
     *      "device_id" : <DEVICE_ID>,
     *      "access_level" : <ACCESS_LEVEL>,
     *      "sharing_level" : <SHARING_LEVEL>,
     *      "email_address" : <EMAIL_ADDRESS>,
     *  }
     *
     * TODO: Add to REST API documentation
     */
    var data map[string]interface{}
    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")

    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    deviceId, ok := data["device_id"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"device_id expected\"}")
        return
    }

    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()

    device, err := dl.LookupDeviceByStringId(deviceId)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"device_lookup_failed\"}");
        return
    }

    //accessLevel, ok := data["access_level"].(int)
    /*_, ok = data["access_level"].(float)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"access_level expected\"}")
        return
    }*/

    //sharingLevel, ok := data["sharing_level"].(int)
    /*_, ok = data["sharing_level"].(float)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"sharing_level expected\"}")
        return
    }*/

    email, ok := data["email"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"email expected\"}")
        return
    }
    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            writeNotLoggedInError(w);
            return
        }
    } else {
        writeNotLoggedInError(w);
        return
    }

    account, err := dl.LookupAccount(username_string)
    if account == nil || err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return
    }

    mailer, err := mail.NewDefaultMailClient()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Failed to initialize mail client\"}")
        return
    }
    mail := mailer.NewMail();
    err = mail.AddTo(email, "")
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Invalid email recipient\"}")
        return
    }
    mail.SetSubject(device.GetFriendlyName())
    mail.SetHTML(`
<img src="http://devel.canopy.link/canopy_logo.jpg"></img>
<h2>I've shared a device with you.</h2>
<a href="http://devel.canopy.link/go.php?share_device=` + deviceId + `">` + device.GetFriendlyName() + `</a>
<h2>What is Canopy?</h2>
<b>Canopy</b> is a secure platform for monitoring and controlling physical
devices.  Learn more at <a href=http://devel.canopy.link>http://canopy.link</a>
`)
    mail.SetFrom("greg@canopy.link", "greg (via Canopy)")
    err = mailer.Send(mail)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"Error sending email\"}")
        return
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\"}");
    return 
}

func finishShareTransactionHandler(w http.ResponseWriter, r *http.Request) {
    /*
     *  POST
     *  {
     *      "device_id" : <DEVICE_ID>,
     *  }
     *
     * TODO: Add to REST API documentation
     * TODO: Highly insecure!!!
     */
    var data map[string]interface{}
    writeStandardHeaders(w);
    session, _ := store.Get(r, "canopy-login-session")

    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&data)
    if err != nil {
        fmt.Fprintf(w, "{\"error\" : \"json_decode_failed\"}")
        return
    }

    deviceId, ok := data["device_id"].(string)
    if !ok {
        fmt.Fprintf(w, "{\"error\" : \"device_id expected\"}")
        return
    }

    var username_string string
    username, ok := session.Values["logged_in_username"]
    if ok {
        username_string, ok = username.(string)
        if !(ok && username_string != "") {
            w.WriteHeader(http.StatusUnauthorized);
            fmt.Fprintf(w, "{\"error\" : \"not_logged_in\"");
            return
        }
    } else {
        w.WriteHeader(http.StatusUnauthorized);
        fmt.Fprintf(w, "{\"error\" : \"not_logged_in\"");
        return
    }

    dl := datalayer.NewCassandraDatalayer()
    dl.Connect("canopy")
    defer dl.Close()
    account, err := dl.LookupAccount(username_string)
    if account == nil || err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"account_lookup_failed\"}");
        return
    }

    device, err := dl.LookupDeviceByStringId(deviceId)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest);
        fmt.Fprintf(w, "{\"error\" : \"device_lookup_failed\"}");
        return
    }

    /* Grant permissions to the user to access the device */
    err = device.SetAccountAccess(account, datalayer.ReadWriteShareAccess)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError);
        fmt.Fprintf(w, "{\"error\" : \"could_not_grant_access\"}");
        return
    }

    fmt.Fprintf(w, "{\"result\" : \"ok\", \"device_friendly_name\" : \"%s\" }", device.GetFriendlyName());
    return 
}

var gPigeon = pigeon.InitPigeonSystem()

var gConfAllowOrigin = ""

func main() {
    fmt.Println("starting server");

    gConfAllowOrigin = os.Getenv("CCS_ALLOW_ORIGIN");
    if (gConfAllowOrigin == "") {
        fmt.Println("You must set environment variable CCS_ALLOW_ORIGIN");
        return
    }

    r := mux.NewRouter()
    r.HandleFunc("/create_account", createAccountHandler)
    r.HandleFunc("/create_device", createDeviceHandler)
    /*r.HandleFunc("/device/{id}", getDeviceInfoHandler).Methods("GET");*/
    r.HandleFunc("/device/{id}", controlHandler).Methods("POST");
    r.HandleFunc("/device/{id}/{sensor}", sensorDataHandler).Methods("GET");
    r.HandleFunc("/devices", devicesHandler)
    r.HandleFunc("/share", shareHandler)
    r.HandleFunc("/finish_share_transaction", finishShareTransactionHandler)
    r.HandleFunc("/login", loginHandler);
    r.HandleFunc("/logout", logoutHandler);
    r.HandleFunc("/me", meHandler);

    http.Handle("/echo", websocket.Handler(CanopyWebsocketServer))
    http.Handle("/", r)
    //err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", context.ClearHandler(http.DefaultServeMux))
    srv := &http.Server{
        Addr: ":8080",
        Handler: context.ClearHandler(http.DefaultServeMux),
        //ReadTimeout: 10*time.Second,
        //WriteTimeout: 10*time.Second,
    }
    err := srv.ListenAndServe()
    if err != nil {
        fmt.Println(err);
    }
}

/*
 * NOTES: Check out https://leanpub.com/gocrypto/read for good intro to crypto.
 */
