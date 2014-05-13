package main

import (
    "canopy/datalayer"
    "flag"
    "fmt"
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
        
    } else {
        fmt.Println("Unknown command: ", flag.Arg(0))
    }
}
