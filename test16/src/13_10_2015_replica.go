package main

import (
	"bufio"
	"io/ioutil"
	"time"
	//"io"
	"net"
	//"os"
	"strconv"
	"strings"
	"sync"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

var (
	requestsPort string //default = 7500
	idx          int
)

func main() {

	idx = 1

	//default port to receive requests
	requestsPort = "7500"

	var wg sync.WaitGroup
	wg.Add(1)

	//listen requests
	go listenRequests(&wg)

	wg.Wait()

}

func listenRequests(localWg *sync.WaitGroup) {
	//get my own IP (introspection?)
	myIP := getMyIP()

	//create the port listener
	ln, _ := net.Listen("tcp", ":"+requestsPort)

	//loop condition for waiting requests
	work := true

	for work {

		//wait connections
		conn, _ := ln.Accept()

		//read the command
		cmd, _ := bufio.NewReader(conn).ReadString('\n')

		s := string(cmd)

		//TODO improve the answer to the client
		//answer := "your command" + s + " was queued\n"
		answer := "ok, cmd= " + s[:len(s)-1] + "\n"

		//answer the client:
		conn.Write([]byte(answer))

		saveMessage("requests executed! :-)")

		//put the command in the queue of etcd

		cfg := client.Config{
			Endpoints: []string{"http://192.168.15.100:4001"},
			Transport: client.DefaultTransport,
			// set timeout per request to fail fast when the target endpoint is unavailable
			HeaderTimeoutPerRequest: time.Second,
		}
		c, err := client.New(cfg)
		if err != nil {
			saveMessage("erro ao criar cliente do etcd")
		} else {
			saveMessage("client of etcd created!")
			kapi := client.NewKeysAPI(c)
			value := myIP + "," + s[:len(s)-1]

			saveMessage("API ready to be called")

			_, err := kapi.CreateInOrder(context.Background(), "/apprequests", value, nil)
			if err != nil {
				//f.WriteString(err)
				saveMessage("erro ao guardar no store etcd")
				saveMessage(err.Error())
				//saveMessage(resp.Action)
			} else {
				saveMessage("all is done, request was scheduled :-)")
			}

			//f.Write([]byte(resp.Action))
			//f.WriteString("response of creating request in the queue: ", resp)

		}
	}
}

func saveMessage(msg string) {
	data := []byte(msg)
	fileName := "/tmp/replicalog-" + strconv.Itoa(idx)
	idx++
	ioutil.WriteFile(fileName, data, 0644)
}

func getMyIP() string {
	ifaces, _ := net.Interfaces()
	//Check(err)

	for _, i := range ifaces {
		addrs, _ := i.Addrs()

		for _, a := range addrs {
			//log.Printf("%v %v\n", i.Name, a)
			var s string
			s = a.String()
			//fmt.Println("a=",s, " part=",s[:6])
			saveMessage("analysing if it is my ip: " + s)
			saveMessage("checking: " + s[:6] + " is equal to 18.16.?")
			//saveMessage("full: " + s[:10])
			if len(s) > 4 && s[:6] == "18.16." { //18.16.xx.y/24 for containers
				//	fmt.Println("This is an address of container:", s[:10])

				//copy until to find '/24'
				parts := strings.Split(s, "/")
				return parts[0]
			}
		}
	}
	return "couldNotDiscoverMyIP"
}
