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
package endpoints

import (
    "canopy/mail"
    "canopy/rest/adapter"
    "canopy/rest/rest_errors"
    "net/http"
)

func POST_share(w http.ResponseWriter, r *http.Request, info adapter.CanopyRestInfo) (map[string]interface{}, rest_errors.CanopyRestError) {
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
    deviceId, ok := info.BodyObj["device_id"].(string)
    if !ok {
        return nil, rest_errors.NewBadInputError("String \"device_id\" expected")
    }

    device, err := info.Conn.LookupDeviceByStringID(deviceId)
    if err != nil {
        return nil, rest_errors.NewBadInputError("Device not found")
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

    email, ok := info.BodyObj["email"].(string)
    if !ok {
        return nil, rest_errors.NewBadInputError("String \"email\" expected")
    }

    if info.Account == nil {
        return nil, rest_errors.NewNotLoggedInError()
    }

    mailer, err := mail.NewDefaultMailClient()
    if err != nil {
        return nil, rest_errors.NewInternalServerError("Error initializing mail client")
    }
    mail := mailer.NewMail();
    err = mail.AddTo(email, "")
    if err != nil {
        return nil, rest_errors.NewBadInputError("Invalid email recipient")
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
        return nil, rest_errors.NewInternalServerError("Error sending mail")
    }

    return map[string]interface{} {
        "result" : "ok",
    }, nil
}
