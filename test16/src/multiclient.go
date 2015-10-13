package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	//"time"
	"strconv"
)

const DEBUG = true

var (

	//declare the the map of IPs
	pods map[string]string

	requestsPort string //default = 7500

	wgClient sync.WaitGroup

	requestsMap map[string]string

	//how many replicas must answer (f+1)?
	majority int
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usange: multiclient R C\nR: how many requests each client should make\nC: how many clients should be instantiated.")
		os.Exit(1)
	}

	//get the number of requests
	nRequests, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(os.Args[1], " is an invalid number of requests: ", err)
		os.Exit(1)
	}

	//get the number of clients
	nClients, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println(os.Args[2], " is an invalid number of clients: ", err)
		os.Exit(1)
	}

	if DEBUG {
		fmt.Printf("clients=", nClients, ", requests/client=", nRequests)
	}

	requestsPort = "7500"

	//TODO get the running PODS of some specific service, and not
	//all pods that are running

	//read the IP's of the running PODS
	resp, err := http.Get("http://192.168.15.100:8080/api/v1/endpoints")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	//read the answer
	contentByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	content := string(contentByte)

	//fmt.Println("IP's retrieved from master: ", content)

	//create the list of PODs
	pods = make(map[string]string)

	//parse the addresses of the PODS
	words := strings.Split(content, "\"ip\":")
	for _, v := range words {
		if v[1:3] == "18" { //18.x.x.x.x
			//get the IP
			ip := v[1:13]
			ultimoCaracter := ip[len(ip)-1 : len(ip)]
			terminaComNumero := strings.Contains("0123456789", ultimoCaracter)
			fmt.Println("ip=", ip, ", ultimoCar=", ultimoCaracter, ", terminaNum=", terminaComNumero)
			for !terminaComNumero {
				//time.Sleep(1 * time.Second)
				ip = ip[:len(ip)-1]
				ultimoCaracter = ip[len(ip)-1 : len(ip)]
				terminaComNumero = strings.Contains("0123456789", ultimoCaracter)
				//fmt.Println("ip=", ip, ", ultimoCar=", ultimoCaracter, ", terminaNum=", terminaComNumero)
			}
			pods[ip] = ip
		}
	}

	//print the set of replicas
	fmt.Print("Replicas: ")
	for _, v := range pods {
		fmt.Print(v, " ")
	}
	fmt.Println("")

	//how much if f+1?
	majority = (len(pods) / 2) + 1

	fmt.Println("majority = ", majority)

	//generate the commands
	requestsMap = make(map[string]string)

	//populate the map of requests
	for i := 0; i < nRequests; i++ {
		nReq := strconv.Itoa(i)
		requestsMap[nReq] = "R" + nReq
	}

	//define how many clients will be working
	wgClient.Add(nClients)

	//start the clients
	for i := 0; i < nClients; i++ {
		go startClient(i, requestsMap, &wgClient)
	}

	//wait all clients finish to make their requests
	wgClient.Wait()
}

func startClient(clientN int, requests map[string]string, localWg *sync.WaitGroup) {

	var wgReplicas sync.WaitGroup

	//create the client ID
	clientID := "C" + strconv.Itoa(clientN)

	//ask the requests
	for _, generalRequest := range requests {

		//prepare the request: add an ID of the client
		request := generalRequest + clientID

		//wait for a majority of answers
		wgReplicas.Add(majority)

		//go through the replicas/PODs
		for _, ipReplica := range pods {

			//ask the replica
			go getAnswerFromReplica(ipReplica, request, &wgReplicas)
		}

		//wait until a majority of replicas answered
		wgReplicas.Wait()
	}

	//this client is done :-)
	wgClient.Done()
}

func getAnswerFromReplica(replicaIP string, command string, localWg *sync.WaitGroup) {

	if DEBUG {
		fmt.Println("connecting in: [", replicaIP, "]")
	}

	//connect to the replica
	conn, err := net.Dial("tcp", replicaIP+":"+requestsPort)
	if err != nil {
		fmt.Println("error connecting to ", replicaIP, ": ", err)
		localWg.Done()
		return
	}

	if DEBUG {
		fmt.Println("connected in ", replicaIP, ", sending get")
	}

	//send the command
	fmt.Fprint(conn, command+"\n")

	if DEBUG {
		fmt.Println("message "+command+" sent to ", replicaIP)
	}

	//receive the answer
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("error reading answer from ", replicaIP, ": ", err)
		localWg.Done()
		return
	}

	if DEBUG {
		fmt.Println("received: [", message[:len(message)-1], "]")
	}

	//the majority of replicas answered, task done :-)
	localWg.Done()
}
