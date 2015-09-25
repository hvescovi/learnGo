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
  //serverID := RandStringRunes(5)
  serverID := "master"

  //get the IP and port of the service
  //server := os.Getenv("REPSVC_SERVICE_HOST")

  //server := "127.0.0.1"

  //if len(server) < 1 {
  //  fmt.Printf("REPSVC_SERVICE_HOST not found, it could be missing in the environment variables definition\n")
  //  return
  //}
  //port := os.Getenv("REPSVC_SERVICE_PORT")
  port:= "8090"
  //if len(port) < 1 {
  //  fmt.Printf("REPSVC_SERVICE_PORT not found, it could be missing in the environment variables definition\n")
  //  return
  //}
  ln, _ := net.Listen("tcp", ":" + port)

  //create a log file
  f, _ := os.Create("/tmp/"+serverID)
  defer f.Close()

  for {
    //get this time
    start := time.Now()

    //wait for conections
    conn, _ := ln.Accept()

    //read the command
    cmd, _ := bufio.NewReader(conn).ReadString('\n')
    s := string(cmd)

    //shutdown the server?
    if s == "shutdown\n" {
      return
    }

    fmt.Print(s," size=")
    fmt.Print(len(s))
    elapsed := time.Since(start).Nanoseconds()

    //prepare the answer
    answer := "server ID = " + serverID + ", answer = "
    answer += strconv.Itoa(int(elapsed))

    //log the answer
    f.WriteString(answer + "\n")

    //send the answer
    conn.Write([]byte(answer+"\n"))
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
