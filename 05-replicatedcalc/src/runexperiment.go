package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	commandMessage = "Specify: n-requests n-clients list-of-nodes-IP(comma separated) port-of-service n-replicas[this, just for logging]"

	nodes []string

	wg sync.WaitGroup
)

func main() {

	if len(os.Args) < 6 {
		fmt.Println(commandMessage)
		os.Exit(0)
	}

	requests, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(commandMessage, ": ", err.Error())
	}
	clients, err2 := strconv.Atoi(os.Args[2])
	if err2 != nil {
		fmt.Println(commandMessage, ": ", err2.Error())
	}
	listOfNodes := os.Args[3]
	port := os.Args[4]

	replicas := os.Args[5]

	nodes = strings.Split(listOfNodes, ",")

	for c := 1; c <= clients; c++ {
		wg.Add(1)
		go client(c, requests, nodes, port, replicas, os.Args[2])
	}

	wg.Wait()
}

func client(clientIntId int, requests int, nodes []string, port string, replicas string, clients string) {
	clientID := "c" + strconv.Itoa(clientIntId)
	opIndex := 1
	operation := ""
	nodeIndex := 0
	maxNode := len(nodes) - 1
	changePort := port == "0" //port of localhost?
	for r := 1; r <= requests; r++ {
		if opIndex == 1 {
			operation = "get"
		} else if opIndex == 2 {
			operation = "inc"
		} else if opIndex == 3 {
			operation = "dou"
			opIndex = 0
		}
		opIndex++
		requestID := clientID + ".r" + strconv.Itoa(r)

		sPort := port
		if changePort {
			nport := 8000 + (r % 3)
			sPort = strconv.Itoa(nport)
		}

		extra := ""
		if port == "0" { //localhost?
			extra = "/" + replicas
		}

		urlCommand := "http://" + nodes[nodeIndex] + ":" + sPort + "/" + requestID + "/" + operation + extra
		//fmt.Println("calling: ", urlCommand)

		tryAgain := true
		timesCounter := 1
		error := ""
		for tryAgain {

			//get the initial time
			t0 := time.Now()

			resp, err := http.Get(urlCommand)
			if err != nil {
				error = " ERROR calling " + urlCommand + ": " + err.Error()
				timesCounter++
			} else {
				defer resp.Body.Close()
				body, err2 := ioutil.ReadAll(resp.Body)

				// get final time
				t1 := time.Now()
				d1 := t1.Sub(t0)

				if err2 != nil {
					error += " ERROR reading answer: " + err.Error()
					timesCounter++
				} else {
					tryAgain = false
					answer := string(body)
					answer = answer[:len(answer)-1] //remove the new line char
					fmt.Print(d1.Nanoseconds())
					fmt.Println(";" + clients + ";" + replicas + ";" + answer + ";" + error)
				}
			}
			if timesCounter > 5 {
				fmt.Println("FAILURE: " + error)
				tryAgain = false
			}
		}
		if nodeIndex == maxNode {
			nodeIndex = 0
		} else {
			nodeIndex++
		}
	}
	wg.Done()
}
