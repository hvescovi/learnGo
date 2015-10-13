package main

import (
	//"bufio"
	"io/ioutil"
	"time"
	//"io"
	"net"
	"net/http"
	//"os"
	"fmt"
	"strconv"
	"strings"
	//"sync"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

const DEBUG = true

var (
	requestsPort string //default = 7500
	idx          int
	myIP         string
	queue        map[string]string
)

func main() {

	//sequential counter for debug messages
	idx = 1

	//get my own IP (introspection?)
	//myIP = getMyIP()
	myIP = "192.168.15.100"

	//default port to receive requests
	requestsPort = "7500"

	retrieveQueueOfRequestsFromEtcd()

	//listen requests
	http.HandleFunc("/", defaultHandler)

	http.ListenAndServe(":"+requestsPort, nil)
}

func retrieveQueueOfRequestsFromEtcd() {

	//create the queue memory structure
	queue = make(map[string]string)

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
		saveMessage("client of etcd created! (for reading queue)")
		kapi := client.NewKeysAPI(c)

		saveMessage("API ready to be called (for reading queue)")

		myOps := client.GetOptions{Recursive: true, Sort: true}
		resp, err := kapi.Get(context.Background(), "/apprequests/", &myOps)
		if err != nil {
			saveMessage("erro ao guardar no store etcd")
			saveMessage(err.Error())
		} else {
			saveMessage("let's read the queue")

			//load the queue
			//fmt.Println("value: ", resp.Node.Value)
			//fmt.Println("action: ", resp.Action)
			all := resp.Node.Nodes
			//fmt.Println("how many children: ", all.Len())
			for _, node := range all {

				//fmt.Println(node.Key, "=", node.Value)
				parts := strings.Split(node.Key, "/")
				fmt.Println(parts[2], "=", node.Value)
				queue[parts[2]] = node.Value
			}

			saveMessage("queue readed")
		}
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		fmt.Fprintf(w, "ok get")
	} else if r.Method == "POST" {

		//fmt.Println("in server: printing the keys and values \n")

		//get the new request
		r.ParseForm()
		for k, v := range r.Form {
			if k == "request" {

				//get the command
				cmd := v[0]

				//process the request
				if DEBUG {
					fmt.Println("in server: the request will be processed:<", cmd, ">")
				}

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
					value := myIP + "," + cmd

					saveMessage("API ready to be called")

					_, err := kapi.CreateInOrder(context.Background(), "/apprequests", value, nil)
					if err != nil {
						//f.WriteString(err)
						saveMessage("erro ao guardar no store etcd")
						saveMessage(err.Error())
						//saveMessage(resp.Action)
					} else {
						saveMessage("all is done, request was scheduled :-)")

						//wait until there are f other instancies of this request in the queue (etcd)
						myOps := client.WatcherOptions{AfterIndex: 0, Recursive: true}
						resp := kapi.Watcher("/apprequests", &myOps)

						//TODO it will be majority
						remainingAnswers := 1

						for remainingAnswers > 0 {
							r, err := resp.Next(context.Background())
							if err != nil {
								fmt.Println("error!", err)
							}
							action := r.Action
							fmt.Println("action: ", action)
							returnedValue := r.Node.Value
							fmt.Println("value: ", returnedValue)

							parts := strings.Split(r.Node.Value, ",")
							newRequestInTheQueue := parts[1]

							fmt.Println(newRequestInTheQueue, "=", cmd, "?")
							if newRequestInTheQueue == cmd {
								remainingAnswers--
								fmt.Println("another command entered in the queue")
							}
						}
					}
				}
			}
			//fmt.Println("in server: \n key:", k, "\n value:", v)
			//s := v[0]
			//fmt.Println("in server: v[0]=", s)
		}
		fmt.Fprintf(w, "ok post \n")
	}
}

func putOrderedValueInRequestsQueueOfEtcd(value string) {

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
		value := myIP + "," + value[:len(value)-1]

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
	}
}

func saveMessage(msg string) {

	if DEBUG {
		fmt.Println(msg)
	} else {
		data := []byte(msg)
		fileName := "/tmp/replicalog-" + strconv.Itoa(idx)
		idx++
		ioutil.WriteFile(fileName, data, 0644)
	}
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

			if len(s) > 4 && s[:6] == "18.16." { //18.16.xx.y/24 for containers
				//copy until to find '/24'
				parts := strings.Split(s, "/")
				return parts[0]
			}
		}
	}
	return "couldNotDiscoverMyIP"
}
