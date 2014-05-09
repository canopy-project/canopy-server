package main

import (
    "canopy/datalayer"
)

func main() {
    dl := datalayer.NewCassandraDatalayer()
    /*dl.PrepDb() */
    dl.Connect()
    dl.StorePropertyValue("abcdef", "cpu", 0.87);
}
