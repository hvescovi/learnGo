package main

import (
  "fmt"
  "net"
  //"strings"
)

func main() {

  ln, _ := net.Listen("tcp", ":8080")
  for {
    conn, _ := ln.Accept()
    var cmd[]byte
    fmt.Fscan(conn, &cmd)
    fmt.Print(string(cmd))
    }
}
