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
)

var (

	//declare the the map of IPs
	pods map[string]string

	requestsPort string //default = 7500

	wg sync.WaitGroup
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please specify a command: \n get\n add n\n mult n")
		os.Exit(1)
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

	//create the map
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

	//get the command
	cmd := os.Args[1]

	if cmd == "add" || cmd == "mult" {
		cmd += " " + os.Args[2]
	}

	//how much if f+1?
	var majority int
	majority = (len(pods) / 2) + 1

	fmt.Println("majority = ", majority)

	wg.Add(majority) //wait for a majority of answers

	for _, v := range pods {
		//fmt.Println("activating ", v)
		go getAnswerFromReplica(v, cmd, &wg)
	}

	wg.Wait()
}

func getAnswerFromReplica(replicaIP string, command string, localWg *sync.WaitGroup) {

	fmt.Println("connecting in: [", replicaIP, "]")

	//connect to the replica
	conn, err := net.Dial("tcp", replicaIP+":"+requestsPort)
	if err != nil {
		fmt.Println("error connecting to ", replicaIP, ": ", err)
		wg.Done()
		return
	}

	fmt.Println("connected in ", replicaIP, ", sending get")

	//send the command
	fmt.Fprint(conn, command+"\n")

	fmt.Println("message "+command+" sent to ", replicaIP)

	//receive the answer
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("error reading answer from ", replicaIP, ": ", err)
		wg.Done()
		return
	}

	fmt.Println("received: [", message[:len(message)-1], "]")
	wg.Done()
}
