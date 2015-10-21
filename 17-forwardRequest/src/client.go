package main

import (
	"net"
	"fmt"
	"bufio"
	"os"
)

func main() {

  fmt.Println("os.args, len=",len(os.Args))

  if len(os.Args) < 2 {
	fmt.Println("Please specify address:port")
	os.Exit(1)
  }

  destiny := os.Args[1]

  fmt.Println("connecting in ",":7500/")
  //connect
  conn, err := net.Dial("tcp", destiny)
  if err != nil {
	fmt.Println(err)
	os.Exit(1)
  }

  fmt.Println("sending command")
  //sending the command
  fmt.Fprintf(conn, "hello please execute my request")

  fmt.Println("reading answer")
  //read the answer
  ans, _ := bufio.NewReader(conn).ReadString('\n')
  s := string(ans)

  fmt.Println("answer: ", s)
}