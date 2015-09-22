package main

import (
  "fmt"
  "net"
)

func main() {
  conn, _ := net.Dial("tcp", "127.0.0.1:8080")
  fmt.Fprintf(conn, ".\n")
}
