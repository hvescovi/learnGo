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

var (
  serverID, requestsPort, followersPort string
  continueRunning bool
 )

const DEBUG = false //if true, print debuging messages

func main() {
  //create a server ID
  serverID = "leader" + RandStringRunes(5)

  //define which port it will listen
  requestsPort = "8091"

  //define port for followers connect to
  //followersPort = "8092"

  //create the channel with thread that listen requests
  requestsChan := make(chan string)

  //start the listener of requests
  go listenRequests (requestsChan)

  //start the listener of followers
  //sock, _ := net.Listen("tcp", ":" + followersPort)

  var newRequest string

  for {
    //wait for connections from followers
    //conn, _ := sock.Accept()

    if DEBUG {fmt.Println("waiting for messages from the channel")}

    //wait for ...?
    newRequest = <- requestsChan

    if DEBUG {fmt.Println("a new requests arrived: "+newRequest)}
  }
}

func listenRequests(notifyChan chan string) {

  //create the port listener
  ln, _ := net.Listen("tcp", ":" + requestsPort)

  //create a log file
  f, _ := os.Create("/tmp/"+serverID)
  defer f.Close()

  //create the SHARED counter!
  counter := 0

  fmt.Println(serverID, " waiting for connections")

  for {

    //wait for conections
    conn, _ := ln.Accept()

    //read the command
    cmd, _ := bufio.NewReader(conn).ReadString('\n')
    s := string(cmd)

    //debug
    fmt.Print("command received: ", cmd)

    //log the command
    //f.WriteString(cmd)

    if DEBUG {fmt.Println("sending the command via channel for followers")}

    //inform the followers about the new command
    notifyChan <- cmd

    //shutdown the server?
    if s == "shutdown\n" {
      return
    }

    //prepare the answer
    answer := "server ID = " + serverID + ", answer = "

    if DEBUG {fmt.Print("size of commando:",len(s))}
    if DEBUG {fmt.Print(", command: ",s)}

    if s[:3] == "add" {
      //get the number to be incremented
      n := s[4:len(s)-1]
      //fmt.Print(" N = ",n)
      answer += "add" + " " + n

      if DEBUG {fmt.Println("answer="+answer)}

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
