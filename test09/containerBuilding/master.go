package main

import (
  "fmt"
  "net"
  "time"
  "os"
  "strconv"
  "bufio"
  "math/rand"
)

func main() {
  //create a server ID
  serverID := "master"

  //define which port it will listen
  port:= "8091"

  //create the port listener
  ln, _ := net.Listen("tcp", ":" + port)

  //create a log file
  f, _ := os.Create("/tmp/"+serverID)
  defer f.Close()

  //create the SHARED counter!
  counter := 0

  for {
    //wait for conections
    conn, _ := ln.Accept()

    //read the command
    cmd, _ := bufio.NewReader(conn).ReadString('\n')
    s := string(cmd)

    //log the command
    f.WriteString(cmd)

    //shutdown the server?
    if s == "shutdown\n" {
      return
    }

    //prepare the answer
    answer := "server ID = " + serverID + ", answer = "

    //for debugging, show the command
    fmt.Print("size of commando:",len(s))
    fmt.Print(", command: ",s)

    if s[:3] == "add" {
      //get the number to be incremented
      n := s[4:len(s)-1]
      //fmt.Print(" N = ",n)
      answer += "add" + " " + n

      //debug
      fmt.Println("answer="+answer)

      //inc the counter!
      i,err := strconv.Atoi(n)
      Check(err)

      //increment the counter!
      counter += i

    } else if s[:3] == "get" {
      s := strconv.Itoa(counter)
      answer += s
    }

    //send the answer
    conn.Write([]byte(answer+"\n"))
  }
}

func Check(err error) {
  if err!=nil {
    fmt.Println("ERROR: ", err)
    os.Exit(0)
  }
}

func init() {
    rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}
