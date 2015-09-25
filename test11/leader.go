package main

import (
  "fmt"
  "net"
  "time"
  "os"
  "strconv"
  "bufio"
  "math/rand"
  "sync"
)

var (
  serverID, requestsPort, followersPort string
  continueRunning bool
  followersAddresses []string
  nFollowers int
)

const DEBUG = true //if true, print debuging messages

func main() {
  var wg sync.WaitGroup
  wg.Add(1) //one threads will decide that the server continues alive

  //create a server ID
  serverID = "leader" + RandStringRunes(5)

  //define which port it will listen
  requestsPort = "8091"

  //define port for followers connect to
  //followersPort = "8092"

  //create the channel with thread that listen requests
  requestsChan := make(chan string)

  //by now, there are now followers
  nFollowers = 0

  //start the listener of requests
  go notifyFollowers(requestsChan)

  //start the listener of followers
  go listenRequests(requestsChan, &wg)

  wg.Wait()
}

func notifyFollowers(notifyChan chan string) {

  var newRequest string

  for {

    if DEBUG {fmt.Println("waiting for messages from the channel")}

    //wait notifications of new requests
    newRequest = <-notifyChan

    if DEBUG {fmt.Print("let's notify followers, a new requests arrived: "+newRequest)}

    for i := 0; i < nFollowers; i++ {

      //connect in the follower
      if DEBUG {fmt.Println("connecting to the follower ",followersAddresses[i])}
      conn, err := net.Dial("tcp", followersAddresses[i])
      Check(err)

      //send the notification to the follower
      fmt.Fprintf(conn, newRequest)

      //close the connection
      if DEBUG {fmt.Println("notification sent, closing connection")}
      conn.Close()
    }
  }
}

func listenRequests(notifyChan chan string, localWg *sync.WaitGroup) {

  //create the port listener
  ln, _ := net.Listen("tcp", ":" + requestsPort)

  //create a log file
  f, _ := os.Create("/tmp/"+serverID)
  defer f.Close()

  //create the SHARED counter!
  counter := 0

  fmt.Println(serverID, " waiting for connections")

  //variable that controls the main loop
  goOut := false

  for ; !goOut; {

    //wait for conections
    conn, _ := ln.Accept()

    //read the command
    cmd, _ := bufio.NewReader(conn).ReadString('\n')
    s := string(cmd)

    //debug
    fmt.Print("command received: ", cmd)

    //log the command
    f.WriteString(cmd)

    if DEBUG {fmt.Println("sending the command via channel for followers")}

    //inform the followers about the new command
    notifyChan <-cmd

    //shutdown the server?
    if s == "shutdown\n" {
      localWg.Done()
      goOut = true
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
