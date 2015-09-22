package main

import (
  "fmt"
  "net"
  "bufio"
  "os"
)

func main() {

  //get server address and port from enviroment variables
  server := os.Getenv("REPSVC_SERVICE_HOST")
  port := os.Getenv("REPSVC_SERVICE_PORT")

  //connect
  conn, _ := net.Dial("tcp", server + ":" + port)

  //send the command
  cmd := os.Args[1]
  if cmd == "inc" {
    fmt.Fprintf(conn, "inc\n")
  } else if cmd == "shutdown" {
    fmt.Fprintf(conn, "shutdown\n")
  } else {
    fmt.Printf("valid commands: \n inc")
    return
  }
  
  //receive the answer
  message, _ := bufio.NewReader(conn).ReadString('\n')
  fmt.Println(message)
}
