package main

import (
    "fmt"
    "log"
    "net"
)

func localAddresses() {
    ifaces, err := net.Interfaces()
    if err != nil {
        log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
        return
    }
    for _, i := range ifaces {
        addrs, err := i.Addrs()
        if err != nil {
            log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
            continue
        }
        for _, a := range addrs {
            log.Printf("%v %v\n", i.Name, a)
            var s string
            s = a.String()
            fmt.Println("a=",s, " part=",s[:6])

            if len(s) >5 && s[:6] == "10.245" {
              fmt.Println("This is an address of container!")

            }
        }
    }
}

func main() {
    localAddresses()
}
