package main

import (
  "fmt"
  "net"
  "bufio"
)

func main() {

  conn, _ := net.Dial("tcp", "127.0.0.1:8080")
  fmt.Fprintf(conn, "vai\n")
  message, _ := bufio.NewReader(conn).ReadString('\n')
  fmt.Println(message)
}
