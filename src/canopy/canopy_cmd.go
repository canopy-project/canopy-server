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
    "github.com/gocql/gocql"
    "canopy/datalayer"
    "canopy/mail"
    "flag"
    "fmt"
    "time"
)

func main() {
    flag.Parse()
    if flag.Arg(0) == "help" {
        fmt.Println("Usage:");
    } else if flag.Arg(0) == "erase-db" {
        dl := datalayer.NewCassandraDatalayer()
        dl.EraseDb("canopy")
    } else if flag.Arg(0) == "create-db" {
        dl := datalayer.NewCassandraDatalayer()
        dl.PrepDb("canopy")
    } else if flag.Arg(0) == "create-account" {
        dl := datalayer.NewCassandraDatalayer()
        dl.Connect("canopy")
        dl.CreateAccount(flag.Arg(1), flag.Arg(2), flag.Arg(3))
    } else if flag.Arg(0) == "delete-account" {
        dl := datalayer.NewCassandraDatalayer()
        dl.Connect("canopy")
        dl.DeleteAccount(flag.Arg(1))
    } else if flag.Arg(0) == "reset-db" {
        dl := datalayer.NewCassandraDatalayer()
        dl.EraseDb("canopy")
        dl.PrepDb("canopy")
    } else if flag.Arg(0) == "create-device" {
        dl := datalayer.NewCassandraDatalayer()
        dl.Connect("canopy")

        account, err := dl.LookupAccount(flag.Arg(1))
        if err != nil {
            fmt.Println("Unable to lookup account ", flag.Arg(1), ":", err)
            return
        }

        device, err := dl.CreateDevice(flag.Arg(2))
        if err != nil {
            fmt.Println("Unable to create device: ", err)
            return
        }

        err = device.SetAccountAccess(account, 4)
        if err != nil {
            fmt.Println("Unable to grant account access to device: ", err)
            return
        }
    } else if flag.Arg(0) == "list-devices" {
        dl := datalayer.NewCassandraDatalayer()
        dl.Connect("canopy")

        account, err := dl.LookupAccount(flag.Arg(1))
        if err != nil {
            fmt.Println("Unable to lookup account ", flag.Arg(1), ":", err)
            return
        }

        devices, err := account.GetDevices()
        if err != nil {
            fmt.Println("Error reading devices: ", err)
            return
        }
        for _, device := range devices {
            fmt.Printf("%s %s\n", device.GetId(), device.GetFriendlyName())
        }
        
    } else if flag.Arg(0) == "gen-fake-sensor-data" {
        dl := datalayer.NewCassandraDatalayer()
        dl.Connect("canopy")
        deviceId, err := gocql.ParseUUID(flag.Arg(1))
        if err != nil {
            fmt.Println("Error parsing UUID: ", flag.Arg(1), ":", err)
            return;
        }
        device, err := dl.LookupDevice(deviceId)
        if err != nil {
            fmt.Println("Device not found: ", flag.Arg(1), ":", err)
            return;
        }
        for i := 0; i < 100; i++ {
            val := float64(i % 16);
            t := time.Now().Add(time.Duration(-i)*time.Second)
            err = device.InsertSensorSample(flag.Arg(2), t, val)
            if err != nil {
                fmt.Println("Error inserting sample: ", err)
            }
        }

    } else if flag.Arg(0) == "test-email" {
        mailer, err := mail.NewDefaultMailClient()
        if err != nil {
            fmt.Println("Error initializing mail client: ", err)
            return
        }
        mail := mailer.NewMail();
        err = mail.AddTo(flag.Arg(1), "Customer")
        if err != nil {
            fmt.Println("Invalid recipient: ", flag.Arg(1), err)
            return
        }
        mail.SetSubject("Test email from Canopy")
        mail.SetHTML("<b>Canopy Rulez</b>")
        mail.SetFrom("no-reply@canopy.link", "The Canopy Team")
        err = mailer.Send(mail)
        if err != nil {
            fmt.Println("Error sending email:", err)
            return
        }
        fmt.Println("Email sent.")
    } else {
        fmt.Println("Unknown command: ", flag.Arg(0))
    }
}
