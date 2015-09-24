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
    fmt.Printf("Please, specify the client address.\n")
      os.Exit(0)
  }

  if len(os.Args) < 3 {
    fmt.Printf("USAGE: client service_address <command>\n"+
      "possible commands:\n"+
      "   add N\n"+
      "   get\n"+
      "   shutdown\n")
      os.Exit(0)
  }

  //get the server address
  server := os.Args[1]

  //connect
  conn, err := net.Dial("tcp", server+":8090")
  Check(err)

  //send the command
  cmd := os.Args[2]
  ok := (cmd[:3] == "add" || cmd == "shutdown" || cmd == "get")
  if cmd[:3] == "add" {
    cmd += " " + os.Args[3]
  }

  if ok {
    //fmt.Println("sending command: ", cmd)
    fmt.Fprintf(conn, cmd+"\n")
  } else {
    fmt.Printf("invalid command: ", cmd)
    return
  }

  //receive the answer
  message, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    fmt.Println(err)
    return
  }
  //print answer without the last character (without the line break)
  fmt.Println(message[:len(message)-1])
}

func Check(err error) {
  if err!=nil {
    fmt.Println("ERROR: ", err)
    os.Exit(0)
  }
}
