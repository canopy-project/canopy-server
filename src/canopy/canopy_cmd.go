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
    } else {
        fmt.Println("Unknown command: ", flag.Arg(0))
    }
}
