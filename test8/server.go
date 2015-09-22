package main

import (
  "fmt"
  "net"
  "time"
  "os"
  "strconv"
  "bufio"
)

func main() {
  server := os.Getenv("KUBERNETES_SERVICE_HOST")
  port := os.Getenv("KUBERNETES_SERVICE_PORT")
  ln, _ := net.Listen("tcp", server + ":" + port)
  for {
    start := time.Now()
    conn, _ := ln.Accept()
    //var cmd[]byte
    //fmt.Fscan(conn, &cmd)
    cmd, _ := bufio.NewReader(conn).ReadString('\n')
    s := string(cmd)
    if s == "shutdown\n" {
      //fmt.Println("shutdown")
      return
    }

    fmt.Print(s," size=")
    fmt.Print(len(s))
    elapsed := time.Since(start).Nanoseconds()
    //fmt.Println(elapsed)
    conn.Write([]byte(strconv.Itoa(int(elapsed))))
    conn.Write([]byte("\n"))
  }
}
