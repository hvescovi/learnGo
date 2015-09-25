package main

import (
  "net"
  "fmt"
  "os"
)

func main() {
  ifaces, err := net.Interfaces()
  Check(err)
  // handle err
  for _, i := range ifaces {
      addrs, err := i.Addrs()
      Check(err)
      for _, addr := range addrs {
          var ip net.IP
          switch v := addr.(type) {
          case *net.IPNet:
                  ip = v.IP
                  fmt.Println("IPNet: ",ip)
          case *net.IPAddr:
                  ip = v.IP
                  fmt.Println("IPAddr: ",ip)
          }
          // process IP address
          
      }
  }
}

func Check(err error) {
  if err!=nil {
    fmt.Println("ERROR: ", err)
    os.Exit(0)
  }
}
