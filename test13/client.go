package main

import (
  "fmt"
  "net"
  "bufio"
  "os"
)

const DEBUG = false // if true, print debug messages

func main() {

  //get server address and port from parateters
  if len(os.Args) < 2 {
    fmt.Printf("Please, specify the client address.\n")
      os.Exit(0)
  }

  if len(os.Args) < 3 {
    fmt.Printf("USAGE: client accessMode ipAddress [port] <command>\n"+
      "accessMode: direct or port (you only have to specify the port is the 'port' mode)\n"+
      "possible commands:\n"+
      "   add N\n"+
      "   get\n"+
      "   sht\n")
      os.Exit(0)
  }

  //access with address or with address:port?
  accessMode := os.Args[1]

  //if port specified, all parameters increase by 1 in position
  incParameter := 0

  destiny := ""
  if accessMode == "direct" {
    destiny = os.Args[2]
  } else if accessMode == "port" {
    destiny = os.Args[2] + ":" + os.Args[3]
    incParameter = 1
  }

  if DEBUG {fmt.Println("destiny="+destiny)}

  //connect
  conn, err := net.Dial("tcp", destiny)
  Check(err)

  //mount the command
  cmd := os.Args[3+incParameter]
  if DEBUG {fmt.Println("command to analyse: "+cmd)}
  ok := (cmd[:3] == "add" || cmd == "sht" || cmd == "get")
  if cmd[:3] == "add" {
    cmd += " " + os.Args[4+incParameter]
  }

  if ok {
    if DEBUG {fmt.Println("sending command: ", cmd)}
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
