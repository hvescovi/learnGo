package main

import (
  "fmt"
  "net"
  "bufio"
  "os"
  "sync"
  "strings"
  "math/rand"
  "strconv"
)

const DEBUG = false // if true, print debug messages
const DEPLOY_IN_KUBERNETES // if false, devmode with parameters
                          // if true, get IP's from kubernetes/apis
var (
  requestsPort string
  values map[string]string
)

func main() {

  var wg sync.WaitGroup
  wg.Add(1) //block the finish of the program while 1 thread alives

  //initialize the values map
  values = make(map[string]string)

  //create a log file
  //f, _ := os.Create("/tmp/test13follower-ip"+getMyIP())
  //defer f.Close()

  //define a port to listen
  requestsPort = "9000"

  //initialize the list of replicas
  replicas = make(map[string]string)

  if ! DEPLOY_IN_KUBERNETES { //DEVMODE

    //get list of IPs if other replicas
    for i, arg := range os.Args {
      //if there is not point in the parameter, it is the requestsPort!
      if !string.Contains(arg, ".") {
        requestsPort = arg
      } else {
        //add the IP to the list of replicas
        //TODO a list could be simpler than a map?
        replicas[arg] = arg
      }
    }
  } else {
    //try to discover the leader service and port from
    //environmental variables
    //TEST13LEADER_SERVICE_HOST=10.247.91.58
    //TEST13LEADER_SERVICE_PORT=8091

    //get the IP of the service
    leaderServiceAddress := os.Getenv("TEST13LEADER_SERVICE_HOST")

    //get the PORT of the service (it should be 8091, but let's ask it anyway)
    leaderServicePort := os.Getenv("TEST13LEADER_SERVICE_PORT")

    //destiny = leaderServiceAddress + ":" + leaderServicePort
  }

  //start waiting for messages from leader
  go listenRequests(followerID, &wg)

  //wait for the thread above
  wg.Wait()
}



}

func listenRequests(myID string, localWg *sync.WaitGroup) {

  //create the port listener
  ln, _ := net.Listen("tcp", ":" + requestsPort)

  //create a log file
  //f, _ := os.Create("/tmp/"+serverID)
  //defer f.Close()

  fmt.Println("replica waiting for requests")

  //variable that controls the main loop
  goOut := false

  for ; !goOut; {

    //wait for conections
    conn, _ := ln.Accept()

    //read the command
    cmd, _ := bufio.NewReader(conn).ReadString('\n')
    s := string(cmd)

    //debug
    fmt.Println("command received: ", cmd)

    //log the command
    //f.WriteString(cmd)

    //shutdown?
    if s == "sht\n" {
      localWg.Done()
      goOut = true

    } else if s == "get\n" {
      answer := values["counter"]

      if DEBUG {fmt.Println("answering some client, counter="+answer)}

      //send the answer back to the client
      conn.Write([]byte(answer+"\n"))

    } else { //it must be an update command like: key=value

      //prepare the answer
      answer := myID + "=>"

      if DEBUG {fmt.Print("size of commando:",len(s))}
      if DEBUG {fmt.Println(", command: ",s)}

      //parse the command in: key=value
      newValue := strings.Split(s,"=")

      //set the new value
      values[newValue[0]] = newValue[1]

      //complete the answer
      answer += "ok value updated"

      if DEBUG {fmt.Println("answer="+answer)}

      //send the answer back to the client
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

func getMyIP() string {
    ifaces, err := net.Interfaces()
    Check(err)

    for _, i := range ifaces {
        addrs, err := i.Addrs()
        if err != nil {
            //log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
            continue
        }
        for _, a := range addrs {
            //log.Printf("%v %v\n", i.Name, a)
            var s string
            s = a.String()
            //fmt.Println("a=",s, " part=",s[:6])

            if len(s) >5 && s[:6] == "10.245" {  //10.246 for containers
              fmt.Println("This is an address of container:",s[:10])
              return s[:10]
            }
        }
    }
    return ""
}
