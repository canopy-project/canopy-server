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
    } else {
        fmt.Println("Unknown command: ", flag.Arg(0))
    }
}
