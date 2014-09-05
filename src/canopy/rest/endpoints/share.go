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

import (
    "canopy/datalayer/cassandra_datalayer"
    "canopy/mail"
    "encoding/json"
    "fmt"
    "net/http"
)

func POST_share(w http.ResponseWriter, r *http.Request) {
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

    dl := cassandra_datalayer.NewDatalayer()
    conn, err := dl.Connect("canopy")
    if err != nil {
        writeDatabaseConnectionError(w)
        return
    }
    defer conn.Close()

    device, err := conn.LookupDeviceByStringID(deviceId)
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

    account, err := conn.LookupAccount(username_string)
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
    mail.SetSubject(device.Name())
    mail.SetHTML(`
<img src="http://devel.canopy.link/canopy_logo.jpg"></img>
<h2>I've shared a device with you.</h2>
<a href="http://devel.canopy.link/go.php?share_device=` + deviceId + `">` + device.Name() + `</a>
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
