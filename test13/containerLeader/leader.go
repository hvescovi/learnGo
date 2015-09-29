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
  followers map[string]string
)

const DEBUG = true //if true, print debuging messages

func main() {
  var wg sync.WaitGroup
  wg.Add(1) //one threads will decide that the server continues alive

  //initializing the map of followers
  followers = make(map[string]string)

  //create a server ID
  serverID = "leader" + RandStringRunes(5)

  //define which port the LEADER will listen
  requestsPort = "8091"

  //define port for followers connect to
  //followersPort = "8092"

  //create the channel with thread that listen requests
  requestsChan := make(chan string)

  //start the listener of followers
  go listenRequests(requestsChan, &wg)

  //start the listener of requests
  go notifyFollowers(requestsChan)

  wg.Wait()
}

func notifyFollowers(notifyChan chan string) {

  var newUpdate string

  for {

    if DEBUG {fmt.Println("waiting for messages from the channel")}

    //wait notifications of new requests
    newUpdate = <-notifyChan

    if DEBUG {fmt.Println("let's notify followers, a new update arrived: "+newUpdate)}

    for name, address := range followers {

      //connect in the follower
      if DEBUG {fmt.Println("connecting to the follower ", name," at "+address)}
      conn, err := net.Dial("tcp", address)
      Check(err)

      //send the notification to the follower
      fmt.Fprintf(conn, newUpdate)

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

    //commands from clients
    cc1 := s[:3] == "add"
    cc2 := s[:3] == "get"
    cc3 := s == "sht\n"
    //commands from followers
    cl1 := s[:3] == "rgt"
    //shutdown the server?
    if cc3 {
      localWg.Done()
      goOut = true

    } else if cc1 || cc2 { //command from client

      //prepare the answer
      answer := serverID + "=>"

      if DEBUG {fmt.Print("size of commando:",len(s))}
      if DEBUG {fmt.Print(", command: ",s)}

      if s[:3] == "add" {
        //get the number to be incremented
        n := s[4:len(s)-1]
        //fmt.Print(" N = ",n)
        answer += "add" + " " + n

        if DEBUG {fmt.Println("answer="+answer)}

        //converting the counter
        i,err := strconv.Atoi(n)
        Check(err)

        //increment the counter!
        counter += i

        //prepare the new value update for the followers
        followerCommand := "counter="
        //convert back the new updated value
        newValue := strconv.Itoa(counter)
        //complete the command for the follower
        followerCommand += newValue

        //inform the followers about the new updating command
        notifyChan <- followerCommand
        if DEBUG {fmt.Println("the new value was sent via channel to the thread that deals with followers")}

      } else if s[:3] == "get" {
        s := strconv.Itoa(counter)
        answer += s
      }
      //send the answer to the client
      conn.Write([]byte(answer+"\n"))

    } else if cl1 { //command from follower

      if DEBUG {fmt.Println("command from follower: "+s)}

      answer := "unknown command"
      if s[:3] == "rgt" {
        //some follower wants to register
        //command: rgt IP:PORT or rgt IP

        //generates a random identification for this follower
        followerID := RandStringRunes(5)

        //get the address
        followerAddress := s[4:len(s)-1]

        //add the follower
        followers[followerID] = followerAddress

        if DEBUG {fmt.Println("follower registered:", followerID, " at ",followerAddress)}

        //answer the random ID generated at leader
        answer = followerID
      }

      //send the answer to the follower
      conn.Write([]byte(answer+"\n"))
    }
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
