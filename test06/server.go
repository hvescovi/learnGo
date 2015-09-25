package main

import "net"
import "fmt"
import "bufio"
//import "strings"

func main() {

  fmt.Println("launching server")

  // listen on all interfaces
  ln, _ := net.Listen("tcp", ":8081")

  // accept connection on port
  conn, _ := ln.Accept()

  // run loop forever
  for {
    message, _ := bufio.NewReader(conn).ReadString('\n')
    fmt.Print("Message received: ", string(message))
    answer := "ok"
    conn.Write([]byte(answer))
  }
}
