package main

import (
  "net"
  "fmt"
  "os"
)

func main() {
  getMyIP()
}

func getMyIP() {
    ifaces, err := net.Interfaces()
    Check(err)

    for _, i := range ifaces {
        addrs, err := i.Addrs()
        if err != nil {
            //log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
            continue
        }
        for _, a := range addrs {
            //log.Printf("%v %v\n", i.Name, a)
            var s string
            s = a.String()
            //fmt.Println("a=",s, " part=",s[:6])

            if len(s) >5 && s[:6] == "10.246" {
              fmt.Println("This is an address of container:",s[:10])
              //addressReturn := s
              //return //addressReturn
            }
        }
    }
}

func Check(err error) {
  if err!=nil {
    fmt.Println("ERROR: ", err)
    os.Exit(0)
  }
}
