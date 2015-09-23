package main

import (
  "fmt"
  "net"
  "bufio"
  "os"
)

func main() {

  //get server address and port from parateters
  if len(os.Args) < 2 {
    fmt.Printf("USAGE: client service_address port <command>\n"+
      "possible commands:\n"+
      "inc\n"+
      "shutdown\n")
      return
  }
  server := os.Args[1]
  port := os.Args[2]

  //connect
  conn, err := net.Dial("tcp", server + ":" + port)
  if err != nil {
    fmt.Println(err)
    return
  }

  //send the command
  cmd := os.Args[3]
  if cmd == "inc" {
    fmt.Fprintf(conn, "inc\n")
  } else if cmd == "shutdown" {
    fmt.Fprintf(conn, "shutdown\n")
  } else {
    fmt.Printf("valid commands: \n inc")
    return
  }

  //receive the answer
  message, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println(message)
}
