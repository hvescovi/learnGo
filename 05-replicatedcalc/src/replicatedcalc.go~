package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"myutils"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafov/bcast"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

var (

	//just change the following variables to modify the app running mode:
	// ******************************************************************

	//runningInKubernetes = false //run in localhost or kubernetes
	runningInKubernetes = true

	//etcdAPIServer = "http://localhost:4001"
	//etcdAPIServer = "http://192.168.15.100:4001" //virtual machines
	//etcdAPIServer = "http://172.31.42.109:4001" //amazon ec2
	//etcdAPIServer = "http://150.162.64.182:4001"  //cluster - real IP's
	etcdAPIServer = "http://192.168.1.200:4001" //cluster - isolated IP's

	//kubernetesAPIServer = "http://localhost:8080"
	//kubernetesAPIServer = "http://192.168.15.100:8080"
	//kubernetesAPIServer = "http://172.31.42.109:8080"
	//kubernetesAPIServer = "http://150.162.64.182:8080"
	kubernetesAPIServer = "http://192.168.1.200:8080"

	// ******************************************************************

	queueKey     = "/apprequests" //key in etcd, that stores requests
	electionsKey = "/appelections"

	mu sync.RWMutex

	health = "" //content about health of the program

	logBuffer bytes.Buffer //log variables
	mylog     *log.Logger

	myIP   = "none"       //introspection: what is my ip?
	kapi   client.KeysAPI //etcd API interface
	etcdOn = false        //Is the etcd client connected?
	debug  = false        //is debug, many messages will be printed
	port   = "8000"       //default port for receiving requests

	queue        = make(map[string]string) //internal queue
	inverseQueue = make(map[string]string) //inverse structure, for make search faster

	valuesOfConsensusInstances   = make(map[int]int)       // consensus instance, result of the executed request
	consensusInstanceOfRequestID = make(map[string]int)    // requestID, order of request
	requestIDofConsensusInstance = make(map[int]string)    //TEMP , just for printing by now
	opOfConsensusInstance        = make(map[int]string)    //TEMP , just for printing by now
	queueReqIdAndRespondentKeys  = make(map[string]string) //for faster counting of replicas that answered
	tmpRequestListWhileLoading   = make(map[string]string)

	replicas         []string                //set of replicas
	requestChan      = make(chan myRequest)  //notify watched requests
	executionChan    = make(chan myRequest)  //notify requests that achieved consensus
	requestsToAnswer = make(map[string]bool) //requestID of requests that I have to answer

	group = bcast.NewGroup()
	//https://github.com/grafov/bcast

	restartWatcherChan = make(chan string) //channel for restart watcher
	lastObservedKey    = ""                //last observer key by watcher
	watcherRestarts    = 0                 //counter of watcher restarts

	//consensusInMemory []string //consensus just for printing

	availableConsensusInstance = 1
	initialValueOfStateMachine = 0

	lastExecutedInstanceOfConsensus = 0
	thisReplicaIsTheLeader          = false

	leaderMonitorOn = false

	requestsWaitingForTheLeader      map[string]string
	requestsWaitingForTheLeaderTimes map[string]time.Time
	electionMessages                 = make(map[string]string)
	leaderView                       = 1 //actual view of the system

	sentMyVote = false

	loadingQueue = false

	whenNewLeaderWasDiscovered time.Time
)

type myRequest struct {
	key               string
	consensusInstance int
	quorumSize        int
	containerIP       string
	requestID         string
	operation         string
	view              int
	//	actualState string
}

func main() {

	mylog = log.New(&logBuffer, "log: ", log.Lshortfile) //log configurations

	if runningInKubernetes { //find IP of POD
		myIP = myutils.GetMyIP("18.16.") //pattern of IPs of PODS in kubernetes
		if myIP[0:6] != "18.16." {
			mylog.Println("main: could not get MY own IP!: ", myIP)
		}
	} else {
		myIP = "127.0.0.1" //IP of localhost
	}

	if !runningInKubernetes { //running in localhost mode?
		if len(os.Args) < 3 { //3 parameters: program(mandatory),port,quorum-size
			fmt.Println("you should specify PORT and QUORUM-SIZE when using in " +
				"localhost mode. Assuming PORt=8000 and QUORUM-SIZE=5")
			//os.Exit(0)
			port = "8000"
			quorumSize := 5
			updateSetOfReplicasInLocalhost(quorumSize)
			if debug {
				mylog.Println("handR: [default] quorum size=", quorumSize)
			}
		} else {
			port = os.Args[1] //change the default port to be the specified one
			quorumSize, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println("ERROR: quorum size must be a number: " + os.Args[2])
				return
			}
			updateSetOfReplicasInLocalhost(quorumSize)
			if debug {
				mylog.Println("handR: quorum size=", quorumSize)
			}
			fmt.Println("quorum-size=", quorumSize)
		}

	} else {
		updateSetOfReplicasInKubernetes()
	}

	requestsWaitingForTheLeader = make(map[string]string)
	requestsWaitingForTheLeaderTimes = make(map[string]time.Time)

	health = "starting replicatedcalc" //first health message

	http.HandleFunc("/healthz", handHealth)
	http.HandleFunc("/logs", handLogs)
	http.HandleFunc("/queue", handQueue)
	http.HandleFunc("/election", handElectionQueue)
	http.HandleFunc("/reset", handResetEtcd)
	http.HandleFunc("/", handRequests)
	http.HandleFunc("/consensus", handListConsensus)

	group = bcast.NewGroup() //broadcast group for consensus notifications

	health += "\n IP discovered: " + myIP //add IP info to the health status
	go executor(executionChan)
	go quorumWatcher(requestChan, executionChan)

	go etcdElectionsWatcher()

	//flag to catch requests notifications while loading queue
	loadingQueue = true
	go etcdWatcher(requestChan)    //listen notifications while loading queue
	loadQueueFromEtcd(requestChan) //retrieve queue in etcd
	loadingQueue = false
	go processRequestsArrivedWhileLoading(requestChan)
	go group.Broadcasting(0)
	go watcherMonitor(restartWatcherChan)

	if !runningInKubernetes {
		fmt.Println("NOT running in kubernetes, server started, port=", port)
	}

	mylog.Println("server ready to start: ", myIP) //initial message in the log
	http.ListenAndServe(":"+port, nil)             //start server
}

func processRequestsArrivedWhileLoading(outChan chan myRequest) {
	if debug {
		mylog.Println("processReqArriveLoad: loading requests that arrived during loading")
	}
	for k, v := range tmpRequestListWhileLoading {

		if debug {
			mylog.Println("processReqArriveLoad: processing: ", k)
		}

		queue[k] = v
		inverseQueue[v] = k

		req := mountRequest(k, v)
		insertReqInSpecificQueue(req.requestID, k)

		//if the view in request is higher than my view, update my view
		if req.view > leaderView {
			leaderView = req.view
		}

		//send to quorumwatcher verify if consensus is possible
		outChan <- req
	}
	//destroy tmp list
	tmpRequestListWhileLoading = nil
	if debug {
		mylog.Println("processReqArriveLoad: queue loaded from temporary requests that arrived during loading")
	}
}

// HANDLERS -----------------------------------------------------------

func handListConsensus(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, myIP, ": consensus in my memory:")
	max := len(valuesOfConsensusInstances)
	for i := 1; i <= max; i++ {
		//for ind, item := range valuesOfConsensusInstances {
		item := valuesOfConsensusInstances[i]
		fmt.Fprint(w, i, " => ", item)
		cmd := requestIDofConsensusInstance[i]
		fmt.Fprint(w, " [", cmd, ",")
		op := opOfConsensusInstance[i]
		fmt.Fprintln(w, op, "]")
	}
}

func handHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, health)
}

func handLogs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, &logBuffer)
}

func handQueue(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, myIP, ": local queue:")
	ordKeys := getOrderedKeysOfLQueue()
	for _, k := range ordKeys {
		mu.RLock()
		value := queue[k]
		mu.RUnlock()
		fmt.Fprintln(w, k, "==", value)
	}
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, myIP, ": queue in etcd:")
	if !etcdOn && !connectEtcdClient() {
		http.Error(w, "ERROR creating etcd client, see logs", 500)
		return
	}
	myGet := client.GetOptions{Recursive: true, Sort: true}
	resp, err := kapi.Get(context.Background(), queueKey, &myGet)
	if err != nil {
		http.Error(w, "ERROR getting queue: "+err.Error(), 500)
		return
	}
	for _, node := range resp.Node.Nodes {
		fmt.Fprintln(w, node.Key, ">>", node.Value)
	}
}

func handElectionQueue(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, myIP, ": local election queue:")
	//ordKeys := getOrderedKeysOfLQueue()
	for k, v := range electionMessages {
		fmt.Fprintln(w, k, "==", v)
	}
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, myIP, ": election queue in etcd:")
	if !etcdOn && !connectEtcdClient() {
		http.Error(w, "ERROR creating etcd client/elections, see logs", 500)
		return
	}
	myGet := client.GetOptions{Recursive: true, Sort: true}
	resp, err := kapi.Get(context.Background(), electionsKey, &myGet)
	if err != nil {
		http.Error(w, "ERROR getting election queue: "+err.Error(), 500)
		return
	}
	for _, node := range resp.Node.Nodes {
		fmt.Fprintln(w, node.Key, ">>", node.Value)
	}
}

func handResetEtcd(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "run this action with only THIS container running")
	if !etcdOn && !connectEtcdClient() {
		http.Error(w, "ERROR creating etcd client, see logs", 500)
		return
	}
	myDelete := client.DeleteOptions{Recursive: true}
	_, err := kapi.Delete(context.Background(), queueKey, &myDelete)
	if err != nil {
		fmt.Fprintln(w, "ERROR deleting in etcd: ", err.Error())
	} else {
		fmt.Fprintln(w, "queue deleted in etcd")
	}
	mySet := client.SetOptions{Dir: true}
	_, err2 := kapi.Set(context.Background(), queueKey+"/", "", &mySet)
	if err2 != nil {
		fmt.Fprintln(w, "ERROR re-creating queue in etcd: ", err2.Error())
	} else {
		fmt.Fprintln(w, "queue re-created in etcd")
	}

	myDelete3 := client.DeleteOptions{Recursive: true}
	_, err3 := kapi.Delete(context.Background(), electionsKey, &myDelete3)
	if err3 != nil {
		fmt.Fprintln(w, "ERROR deleting in etcd/elections: ", err3.Error())
	} else {
		fmt.Fprintln(w, "elections deleted in etcd")
	}
	mySet4 := client.SetOptions{Dir: true}
	_, err4 := kapi.Set(context.Background(), electionsKey+"/", "", &mySet4)
	if err4 != nil {
		fmt.Fprintln(w, "ERROR re-creating elections in etcd: ", err4.Error())
	} else {
		fmt.Fprintln(w, "elections re-created in etcd")
	}

	fmt.Fprintln(w, "End of reset. PLEASE, STOP SERVERS AND START THEM AGAIN!")
}

func startElections() {
	//put in etcd proposal to view change
	proposal := "VIEWCHANGE," + strconv.Itoa(leaderView+1) + "," + myIP + ":" + port
	_, contains := electionMessages[proposal]
	if debug {
		mylog.Println("startelec: start election? contains=", contains, ", sentMyVote=", sentMyVote)
	}
	if !contains && !sentMyVote {
		putViewChangeInEtcd(proposal)
		sentMyVote = true
		if debug {
			mylog.Println("startelec: new election startec, newview=", leaderView+1)
		}
	}
}

func electionInProgress() bool {
	//is there any proposal for a next view?
	for _, v := range electionMessages {
		//VIEWCHANGE, n, myIP:port
		parts := strings.Split(v, ",")
		v, _ := strconv.Atoi(parts[1])
		if v == (leaderView + 1) {
			return true
		}
	}
	return false
}

func leaderMonitor() {

	//wait some random time to start
	rand.Seed(time.Now().Unix())
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

	if debug {
		mylog.Println("leaderMonitor started")
	}
	loop := true
	for loop {
		//time.Sleep(15 * time.Second)

		//check if I know all requests that are waiting for the leader actuation
		for k := range requestsWaitingForTheLeader {
			//check if the request has instance order defined
			_, contains := consensusInstanceOfRequestID[k]
			if !contains {

				//how many time this request is waiting?
				t0 := requestsWaitingForTheLeaderTimes[k]
				t1 := time.Now()
				d1 := t1.Sub(t0)
				if d1.Seconds() > 1 {
					//checkif there is any election for the new view in progress
					if !electionInProgress() {
						if debug {
							mylog.Println("leaderMonitor: request ", k, " has no order for more than 1 sec.")
						}
						t2 := time.Now()
						d2 := t2.Sub(whenNewLeaderWasDiscovered)
						if d2.Seconds() > 1 {
							if debug {
								mylog.Println("leaderMonitor: request ", k, " has no order yet, and last election was far. STARTING ANOTHER LEADER!!")
							}
							startElections()
							leaderMonitorOn = false //turn off this monitor
							break
						} else {
							if debug {
								mylog.Println("leaderMonitor: leader was proclamed recently")
							}
						}
					} else {
						if debug {
							mylog.Println("leaderMonitor: election in progress, I will wait a little...")
						}
						time.Sleep(3 * time.Second)
					}
				} else {
					if debug {
						mylog.Println("leaderMon: request ", k, " has been waiting during: ", d1.Seconds(), " seconds. Waiting more...")
					}
					time.Sleep(100 * time.Millisecond)
				}
			}

		}
		loop = (len(requestsWaitingForTheLeader) > 0) && leaderMonitorOn
		time.Sleep(1000)
	}
	leaderMonitorOn = false //turn off this monitor
	if debug {
		mylog.Println("leaderMonitor finished")
	}
}

func handRequests(w http.ResponseWriter, r *http.Request) {

	t0 := time.Now()

	s := r.URL.String()    //get the URL
	u, err := url.Parse(s) //try to parse the URL
	if err != nil {
		http.Error(w, "ERROR parsing URL: "+err.Error(), 500)
		return
	}
	if debug {
		mylog.Println("handR: request arrives: ", u.Path)
	}
	parts := strings.Split(u.Path, "/") //split parts of the URL
	if (len(parts) != 3) && (runningInKubernetes) {
		http.Error(w, "USAGE in kubernetes: http://ip:port/id/operation", 500)
		return
	}
	if (len(parts) != 4) && (!runningInKubernetes) {
		http.Error(w, "In localhost: ip:port/id/operation/quorum-size", 500)
		return
	}
	if runningInKubernetes {
		updateSetOfReplicasInKubernetes()
	} else {
		quorumSize := len(replicas)
		if debug {
			mylog.Println("handR: quorum size informed by the request=", quorumSize)
		}
	}

	fmt.Fprint(w, parts[1]+"/"+parts[2]+";"+myIP)

	t1 := time.Now()
	d1 := t1.Sub(t0)
	fmt.Fprint(w, ";")
	fmt.Fprint(w, d1.Nanoseconds())

	requestID := parts[1] //get requestID from URL
	operation := parts[2] //get operation from URL

	if debug {
		mylog.Println("handR: request received, ", requestID)
	}

	cID, contains := consensusInstanceOfRequestID[requestID]
	if contains { //already executed request?

		//REPEATED CODE

		t1 := time.Now()
		d1 := t1.Sub(t0)
		fmt.Fprint(w, ";")
		fmt.Fprint(w, d1.Nanoseconds())
		fmt.Fprint(w, ";")
		fmt.Fprint(w, d1.Nanoseconds())
		fmt.Fprint(w, ";")
		fmt.Fprint(w, d1.Nanoseconds())

		fmt.Fprint(w, ";")
		fmt.Fprint(w, cID, ";")
		fmt.Fprint(w, strconv.Itoa(valuesOfConsensusInstances[cID]))
		fmt.Fprintln(w, "")

		//fmt.Fprint(w, strconv.Itoa(valuesOfConsensusInstances[cID]))
		return
	}

	requestsToAnswer[requestID] = true //I have to answer this request!

	numberOfReplicas := len(replicas)

	newInstance := 0
	if thisReplicaIsTheLeader || numberOfReplicas == 1 {
		mu.Lock()
		newInstance = availableConsensusInstance
		//update next available consensus
		availableConsensusInstance++
		mu.Unlock()
	} else {
		//add this request in a list that is waiting for the leader
		//note: newInstance = 0
		requestsWaitingForTheLeader[requestID] = createRequestValue(newInstance,
			len(replicas), myIP+":"+port, requestID, operation, leaderView)
		requestsWaitingForTheLeaderTimes[requestID] = time.Now()

		//start a timer to monitors the leader actuation
		if !leaderMonitorOn {
			go leaderMonitor()
			leaderMonitorOn = true
		}
	}

	requestValue := createRequestValue(newInstance,
		len(replicas), myIP+":"+port, requestID, operation, leaderView)

	if debug {
		mylog.Println("handR: value created: ", requestValue)
	}

	//join the broadcast consensus group
	member1 := group.Join()
	//it is important to join the group BEFORE save the value in ETCD,
	//because consensus observer can notify this thread before it achieve
	//the point of receiving the message, and that way it could loss the
	//notification of consensus

	t2 := time.Now()
	d2 := t2.Sub(t1)
	fmt.Fprint(w, ";")
	fmt.Fprint(w, d2.Nanoseconds())

	putReqInEtcd(requestValue) //put request in etcd and get the new ID

	t3 := time.Now()
	d3 := t3.Sub(t2)
	fmt.Fprint(w, ";")
	fmt.Fprint(w, d3.Nanoseconds())

	if debug {
		mylog.Println("handR: request stored (etcd and local), ", requestID)
	}

	solved := numberOfReplicas < 2 //only one replica?
	if debug {
		mylog.Println("handR: replicas=", len(replicas), ", solved=", solved)
	}

	for !solved {
		if debug {
			mylog.Println("handR: waiting watcher notification for ", requestID)
		}
		//wait notification from quorum watcher
		//expecting: requestID of an executed request
		executedRequest := member1.Recv()

		if debug {
			mylog.Println("handR: is is solved? ", executedRequest, "==", requestID)
		}
		solved = (executedRequest == requestID) //the request that I have to answer was solved?
	}

	if !(thisReplicaIsTheLeader || numberOfReplicas == 1) {
		newInstance = consensusInstanceOfRequestID[requestID]
	}

	//leave the notification group
	member1.Close()

	t4 := time.Now()
	d4 := t4.Sub(t3)
	fmt.Fprint(w, ";")
	fmt.Fprint(w, d4.Nanoseconds())

	if numberOfReplicas == 1 { //only one replica operating?
		//TODO this can present problem when executing system with 1 replica
		//and suddenly increase the quorum of replicas

		previousState := 0
		if newInstance == 1 {
			previousState = initialValueOfStateMachine
		} else {
			//get the state of the previous state

			_, contains := valuesOfConsensusInstances[newInstance-1]
			for !contains {
				time.Sleep(1 * time.Millisecond)
				//mylog.Println("handR: just 1 replica, waiting answer of previous instance; newinstance=", newInstance)
				_, contains = valuesOfConsensusInstances[newInstance-1]
			}
			previousState = valuesOfConsensusInstances[newInstance-1]
		}

		//execute request
		result := executeRequest(previousState, operation)
		if debug {
			mylog.Println("handR: only ONE replica; result of ", operation, ":", result)
		}
		lastExecutedInstanceOfConsensus = newInstance
		valuesOfConsensusInstances[newInstance] = result
		consensusInstanceOfRequestID[requestID] = newInstance
		requestIDofConsensusInstance[newInstance] = requestID
		opOfConsensusInstance[newInstance] = operation

		//update last executed state
		lastExecutedInstanceOfConsensus = newInstance

		//commands in memory are not more necessary

		key := inverseQueue[requestValue]
		//req := mountRequest(key, value)
		delete(queue, key)
		delete(inverseQueue, requestValue)
	}

	//get the actual state when this request achieved consensus
	valueAfterExecution := valuesOfConsensusInstances[newInstance]
	fmt.Fprint(w, ";")
	fmt.Fprint(w, newInstance, ";")
	fmt.Fprint(w, strconv.Itoa(valueAfterExecution))
	fmt.Fprintln(w, "")

	//remove the request from the control structure
	//some problem happened when removing these data :-/
	delete(requestsToAnswer, requestID)
}

// QUEUE --------------

func getOrderedKeysOfLQueue() []string {
	var keys []string
	mu.RLock()
	for k := range queue {
		keys = append(keys, k)
	}
	mu.RUnlock()
	sort.Strings(keys)
	return keys
}

func countInQueue(reqID string) (int, string) { //how many times reqID appears in LQueue?
	//return's also the keys of the counted IDs

	counter := 0
	mu.RLock()
	s := ""
	for k, v := range queue {
		//if debug{mylog.Println("countInQ: k=, ",k,",v=",v,",",reqID)}
		parts := mountRequest(k, v)
		if debug {
			mylog.Println("countInQ: request mounted, asking if ", parts.requestID, "=", reqID, "; req=", v)
		}
		if parts.requestID == reqID {
			s = s + k + " "
			counter++
		}
	}
	mu.RUnlock()
	return counter, s
}

func countInQueueTurbo(reqID string) (int, string) {
	//return how many times a requests occurs,
	//and from which replicas they were appeared
	s, contains := queueReqIdAndRespondentKeys[reqID]
	if contains {
		parts := strings.Split(s, " ")
		return len(parts), s
	} else {
		return 0, ""
	}
}

func requestInLQueueLookOnlyValue(value string) bool {
	_, contains := inverseQueue[value]
	return contains
}

func executedRequest(consensusInstance int) bool {
	_, contains := valuesOfConsensusInstances[consensusInstance]
	return contains
}

// ETCD ----------------

func connectEtcdClient() bool {
	cfg := client.Config{Endpoints: []string{etcdAPIServer},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second}
	c, err := client.New(cfg)
	if err != nil {
		mylog.Println("ERROR creating etcd client: ", err.Error())
		etcdOn = false
	} else {
		kapi = client.NewKeysAPI(c)
		mylog.Println("etcd client connected")
		etcdOn = true
	}
	return etcdOn
}

func loadQueueFromEtcd(outChan chan myRequest) {
	if !etcdOn && !connectEtcdClient() {
		mylog.Println("ERROR creating etcd client, see logs")
		return
	}
	myGet := client.GetOptions{Recursive: true, Sort: true}
	resp, err := kapi.Get(context.Background(), queueKey+"/", &myGet)
	if err != nil {
		mylog.Println("ERROR getting queue from etcd: ", err.Error())
		return
	}
	if debug {
		mylog.Println("LOCKING queue")
	}
	queue = nil
	queue = make(map[string]string) //re-create local queue
	inverseQueue = nil
	inverseQueue = make(map[string]string)
	queueReqIdAndRespondentKeys = nil
	queueReqIdAndRespondentKeys = make(map[string]string)
	higherCI := 0 //last used consensus instance in the queue
	for _, node := range resp.Node.Nodes {
		queue[node.Key] = node.Value
		inverseQueue[node.Value] = node.Key

		//mount a request
		//fmt.Println("mounting request from: ", node.Key, " => ", node.Value)
		req := mountRequest(node.Key, node.Value)
		//fmt.Println("inserting ", req.requestID, " in specific queue")
		insertReqInSpecificQueue(req.requestID, node.Key)
		//fmt.Println(req.requestID, " inserted")

		//if the view in request is higher than my view, update my view
		if req.view > leaderView {
			leaderView = req.view
		}

		//it is importante to read the last available consensus instance executed,
		//because when running with replica=1 this value will be used

		if req.consensusInstance > higherCI {
			higherCI = req.consensusInstance
		}

		//send to quorumwatcher verify if consensus is possible
		outChan <- req
	}
	availableConsensusInstance = higherCI + 1
	if debug {
		mylog.Println("queue loaded from etcd")
	}
}

func insertReqInSpecificQueue(req string, etcdKey string) {
	_, contains := queueReqIdAndRespondentKeys[req]
	if contains {
		queueReqIdAndRespondentKeys[req] += " " + etcdKey
	} else {
		queueReqIdAndRespondentKeys[req] = etcdKey
	}
}

func putReqInEtcd(value string) {
	if !etcdOn && !connectEtcdClient() {
		mylog.Println("ERROR creating etcd client, see logs")
		return
	}

	tmp := mountRequest("tmp", value)
	if tmp.consensusInstance != 0 {

		//as notification of etcd is so fast, first put value in reverseQueue.
		//That way, other threads can search in the reverseQueue to know
		//if this request is being treated in this program.
		inverseQueue[value] = "being acquired"
	}
	if debug {
		mylog.Println("putreq: I will save", value)
	}
	resp, err := kapi.CreateInOrder(context.Background(), queueKey, value, nil)
	if err != nil {
		mylog.Println("ERROR creating value in etcd: ", err.Error())
		return
	}
	if tmp.consensusInstance != 0 {
		inverseQueue[value] = resp.Node.Key
		queue[resp.Node.Key] = value //HARDCODED!!
		//req := mountRequest(resp.Node.Key, value)
		insertReqInSpecificQueue(tmp.requestID, tmp.containerIP)
	}

	if debug {
		mylog.Println("putReq: created in etcd and stored in lqueue, " + resp.Node.Key + " for " + value)
	}
}

func putViewChangeInEtcd(value string) {
	if !etcdOn && !connectEtcdClient() {
		mylog.Println("ERROR creating etcd client/election, see logs")
		return
	}
	resp, err := kapi.CreateInOrder(context.Background(), electionsKey, value, nil)
	if err != nil {
		mylog.Println("ERROR creating value/election in etcd: ", err.Error())
		return
	}
	if debug {
		mylog.Println("putVWC: created in etcd, ", resp.Node.Key, " for ", value)
	}
	electionMessages[resp.Node.Key] = value //HARDCODED!!
	if debug {
		mylog.Println("putVWC: stored in lqueue, ", resp.Node.Key, " for ", value)
	}
}

func loadAllElectionMessagesFromEtcd() {
	if !etcdOn && !connectEtcdClient() {
		mylog.Println("ERROR creating etcd client/elections, see logs")
		return
	}
	myGet := client.GetOptions{Recursive: true, Sort: true}
	resp, err := kapi.Get(context.Background(), electionsKey+"/", &myGet)
	if err != nil {
		mylog.Println("ERROR getting queue from etcd: ", err.Error())
		return
	}
	electionMessages = nil
	electionMessages = make(map[string]string)
	for _, node := range resp.Node.Nodes {
		electionMessages[node.Key] = node.Value
	}
	if debug {
		mylog.Println("election messages loaded from etcd")
	}
}

// KUBERNETES ----------------

func updateSetOfReplicasInKubernetes() {

	//TODO get PODS of the service, instead of all PODS
	resp, err := http.Get(kubernetesAPIServer + "/api/v1/endpoints")
	if err != nil {
		mylog.Println("ERROR getting endpoints in kubernetes API: ", err.Error())
		return
	}
	defer resp.Body.Close()
	contentByte, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		mylog.Println("ERROR reading data from endpoints: ", err2.Error())
		return
	}
	content := string(contentByte)

	if debug {
		mylog.Print("updateSetOfrep: ")
	}

	replicas = nil

	words := strings.Split(content, "\"ip\":")
	for _, v := range words {
		if v[1:7] == "18.16." { //18.x.x.x, IP of PODS
			parts := strings.Split(v, ",")
			theIP := parts[0]
			theIP = theIP[1 : len(theIP)-1] //remove " chars from IP
			replicas = append(replicas, theIP)
			if debug {
				mylog.Print("[", theIP, "]")
			}
		}
	}
	if debug {
		mylog.Println("")
	}
}

func updateSetOfReplicasInLocalhost(quorumSize int) {
	replicas = nil //delete all replicas of the set
	defPort, _ := strconv.Atoi(port)
	for i := 0; i < quorumSize; i++ {
		port := defPort + i //add 'i' to the port (ex. 8001, 80002, ...)
		ps := strconv.Itoa(port)
		replicas = append(replicas, myIP+":"+ps) //add IP and port
	}
	if debug {
		mylog.Println("updateSetRepLocal: replicas has size: ", len(replicas))
	}
}

// REQUESTS -----------------

func mountRequest(key string, value string) myRequest { //mount a complete request
	/*
		   key               string
		   consensusInstance int
		   quorumSize        int
		   containerIP       string
		   requestID         string
		   operation         string
			 view int
	*/
	//fmt.Println("mounting request: ", value)
	parts := strings.Split(value, ",")

	consensusInstance, err1 := strconv.Atoi(parts[0])
	if err1 != nil {
		mylog.Println("mountRequest: error getting consensus instance: ", err1.Error())
		consensusInstance = 0
	}
	quorum, err2 := strconv.Atoi(parts[1])
	if err2 != nil {
		mylog.Println("mountRequest: error getting quorum: ", err2.Error())
		quorum = 0
	}
	view, err3 := strconv.Atoi(parts[5])
	if err3 != nil {
		mylog.Println("mountRequest: error getting view: ", err3.Error())
		view = 0
	}
	//fmt.Println("request mounted: ", value)

	return myRequest{key, consensusInstance, quorum, parts[2], parts[3], parts[4], view}
}

func mountRequestValue(request myRequest) string {
	return strconv.Itoa(request.consensusInstance) + "," + strconv.Itoa(request.quorumSize) + "," + request.containerIP + "," +
		request.requestID + "," + request.operation + "," + strconv.Itoa(request.view) //+ "," + request.actualState
}

func createRequestValue(consensusOrder int, quorumSize int,
	ip string, reqID string, op string, view int) string {
	return strconv.Itoa(consensusOrder) + "," + strconv.Itoa(quorumSize) + "," +
		ip + "," + reqID + "," + op + "," + strconv.Itoa(leaderView)
}

func executor(inExecutionChan chan myRequest) {

	//var waitQueueRun []Request
	var waitQueueRun map[int]myRequest
	waitQueueRun = make(map[int]myRequest)

	for {

		if debug {
			mylog.Println("executor: waiting notifications")
		}

		//receive some request that is ready for execution
		req := <-inExecutionChan

		if debug {
			mylog.Println("executor: received: ", req.requestID, "; checking: is ", req.consensusInstance, "==", lastExecutedInstanceOfConsensus+1, "?")
		}

		//put in a queue of waiting requests
		waitQueueRun[req.consensusInstance] = req

		//try to execute as much requests as possible

		continueExecuting := true //let's run at least once
		for continueExecuting {

			continueExecuting = false //avoid infinite loop

			//are there requests in the execution queue?
			if len(waitQueueRun) > 0 {

				//get the request with the lower consensusInstance
				indexLower := getIndexOfRequestWithLowerConsensusInstanceOrder(waitQueueRun)

				//by some strange factor, is it request already executed?
				if indexLower <= lastExecutedInstanceOfConsensus {
					if debug {
						mylog.Println("executor: this request was already executed, deleting and looping: ", waitQueueRun[indexLower].requestID)
					}
					//remove this request!
					delete(waitQueueRun, indexLower)
					//loop!
					continueExecuting = true
				} else {

					if debug {
						mylog.Println("executor: looping, ", waitQueueRun[indexLower].requestID, ", is ", indexLower, "==", lastExecutedInstanceOfConsensus+1, "?")
					}

					//is this one the next to be executed?
					if (indexLower != -1) && (indexLower == (lastExecutedInstanceOfConsensus + 1)) {

						//get the request
						req = waitQueueRun[indexLower]
						//remove from nexties execution
						delete(waitQueueRun, indexLower)
						//let's loop!continueExecuting = true
						continueExecuting = true

						if debug {
							mylog.Println("executor: let's EXECUTE: ", req.requestID, ",order=", req.consensusInstance)
						}

						//get the previous executed state
						oldValue := 0
						if lastExecutedInstanceOfConsensus != 0 {
							oldValue = valuesOfConsensusInstances[req.consensusInstance-1]
							if debug {
								mylog.Println("quorumWatch: old value retrieved: ", oldValue, ",reqID=", req.requestID)
							}
						}

						//execute the request!
						newValue := executeRequest(oldValue, req.operation)

						//save the new value
						lastExecutedInstanceOfConsensus = req.consensusInstance
						valuesOfConsensusInstances[req.consensusInstance] = newValue
						consensusInstanceOfRequestID[req.requestID] = req.consensusInstance
						requestIDofConsensusInstance[req.consensusInstance] = req.requestID
						opOfConsensusInstance[req.consensusInstance] = req.operation

						//I have to inform this consensus to another threads?
						//only if I an the executor of this request
						if isMyResponsabilityToAnswer(req.requestID) {

							//send the notification for threads that are waiting to answer client
							member1 := group.Join()
							if debug {
								mylog.Println("consensus: notifying group", req.requestID, ":", req.containerIP)
							}
							member1.Send(req.requestID)

							if debug {
								mylog.Println("consensus notified for respond the client: ", req.requestID)
							}
						} else {
							if debug {
								mylog.Println("consensus: I do not have to notify for: ", req.requestID)
							}
						}
					}
				}
			}
		}
	}
}

func getIndexOfRequestWithLowerConsensusInstanceOrder(mapOfRequests map[int]myRequest) int {
	if len(mapOfRequests) == 0 {
		return -1
	}
	lower := 0
	for k := range mapOfRequests { //only indexes will be performed
		if lower == 0 {
			lower = k
		} else {
			if k < lower {
				lower = k
			}
		}
	}
	return lower
}

func quorumWatcher(inRequestChan chan myRequest, outExecutionChan chan myRequest) {
	for {
		if debug {
			mylog.Println("consensus watcher waiting...")
		}
		req := <-inRequestChan //new request from etcd watcher!

		if debug {
			mylog.Println("quorumWatcher: starting consensus checking for: ", req.key, ":", req.requestID)
		}

		qs := req.quorumSize //how was quorum size at that time?

		majority := (qs / 2) + 1
		if debug {
			mylog.Println("quorumWatcher: counting in queue: ", req.key, ":", req.requestID)
		}
		//already, tmpQuorumKeys := countInQueue(req.requestID)
		already, tmpKeysToRemove := countInQueueTurbo(req.requestID)

		if debug {
			mylog.Println("quorumWatcher: counting finishedin, there are: ", already, " in queue, of ", req.requestID)
		}
		enoughAnswers := already >= majority

		if enoughAnswers {
			if debug {
				mylog.Println("quorum ACHIEVED: ", req.requestID, ", maj=", majority, ",alred=", already, ", notifying executor...")
			}

			//notify that executor can run this request
			executionChan <- req

			//remove keys that are no more necessary
			go func() {
				parts := strings.Split(tmpKeysToRemove, " ")
				for _, key := range parts {
					value := queue[key]
					//req := mountRequest(key, value)
					delete(queue, key)
					delete(inverseQueue, value)
					//delete(queueReqIdAndReplicas, req.requestID)
				}
			}()

		} else {
			if debug {
				mylog.Println("quorumWatcher: consensus not achieved yet for: ", req.key, ":", req.requestID, "maj=", majority, ", alr=", already)
			}
		}
	}
}

func isMyResponsabilityToAnswer(reqID string) bool {
	_, contains := requestsToAnswer[reqID]
	return contains
}

func queueContainsKey(key string) bool {
	//http://stackoverflow.com/questions/2050391/how-to-check-if-a-map-contains-a-key-in-go
	_, contains := queue[key]
	return contains
}

func getElectionsVotes() string {
	s := ""
	for k := range electionMessages {
		s += k + " "
	}
	return s
}

func getAddressOfLowerKey() string {
	//pre-condition: election messages were loaded from etcd
	lower := new(big.Int)
	lower.SetString("99999999999999999999", 10)
	/*
		lower, err := strconv.ParseUint("99999999999999999999", 10, 64)
		if err != nil {
			msg := "Could not convert initial string  99999...: " + err.Error()
			fmt.Println(msg)
			return msg
		}
	*/

	lowerKey := ""
	for k, v := range electionMessages {
		//fmt.Println("doing for, k=", k)
		nParts := strings.Split(k, "/")
		// format: /key/0000000001234
		numS := nParts[2]
		num := new(big.Int)
		num.SetString(numS, 10)
		//num, err2 := strconv.ParseUint(numS, 10, 64)
		/*
			if err2 != nil {
				msg2 := "Could not convert " + numS + ": " + err.Error()
				fmt.Println(msg2)
				mylog.Println(msg2)
				return msg2
			} else {
				fmt.Println("conversion ok, num=", num)
			}
		*/
		partsV := strings.Split(v, ",")
		//VIEWCHANGE, view, IP:port
		view, _ := strconv.Atoi(partsV[1])

		//fmt.Println("geAddLowKey: checking, num=", num, " < lower=", lower, "? view=", view, "==leaderView=", leaderView, "? extra: value="+v+", key="+k)
		//myDebug("geAddLowKey: checking, lowerKey=" + lowerKey + ", value=" + v + ", key=" + k)
		if num.Cmp(lower) == -1 && view == leaderView {
			lower = num
			lowerKey = k
			myDebug("getaddlowkey: lowerkey=" + k)
		} else {
			myDebug("getaddlowkey: " + k + " is higher than " + lowerKey)
		}
	}
	//VIEWCHANGE, n, myIP:port
	parts := strings.Split(electionMessages[lowerKey], ",")
	return parts[2]
	//return "localhost:8000"
}

func myDebug(message string) {
	if debug {
		mylog.Println(message)
	}
}

func countVotesOfView(view int) int {
	result := 0
	for _, v := range electionMessages {
		parts := strings.Split(v, ",") //VIEWCHANGE,n,myIP:port
		n, _ := strconv.Atoi(parts[1])
		if n == view {
			result++
		}
	}
	return result
}

func etcdElectionsWatcher() {

	mylog.Println("electionsWatcher started")
	if !etcdOn && !connectEtcdClient() {
		mylog.Println("ERROR creating etcd client/elections")
		return
	}
	myWatcher := client.WatcherOptions{AfterIndex: 0, Recursive: true}
	resp := kapi.Watcher(electionsKey, &myWatcher)
	for {

		myDebug("etcdElectionsWatcher: waiting request from watcher...")

		r, err := resp.Next(context.Background())
		if err != nil {
			mylog.Println("etcdElectionsWatcher: ERROR watching etcd: ", err.Error())
			return
		}
		//process election entry
		go func() { //make the cycle of receiving notifications faster!! make it a Go routine!

			if r.Action == "create" {

				parts := strings.Split(r.Node.Value, ",") //VIEWCHANGE,n,myIP:port
				newView, _ := strconv.Atoi(parts[1])
				votingContainer := parts[2]

				myDebug("etcdElectionsWatcher: new message: " + r.Node.Value)

				if !thisReplicaIsTheLeader && (newView == leaderView+1) {

					if debug {
						mylog.Println("etcdElectionsWatcher: vote: ", votingContainer)
					}

					_, has := electionMessages[r.Node.Key]
					if !has {
						//insert the vote in the set of values
						electionMessages[r.Node.Key] = r.Node.Value
					}

					//check if there are enough votes
					votesForThisView := countVotesOfView(newView)
					if debug {
						mylog.Println("etcdElecWatch: votes for view ", newView, ": ", votesForThisView)
					}
					if votesForThisView >= (len(replicas)/2 + 1) {
						//if len(electionMessages) >= (len(replicas)/2 + 1) {

						//UPDATE the view
						//THIS impacts on next commands, since they will use the new view in  comparisons!!
						leaderView++

						//read all elections buffer
						loadAllElectionMessagesFromEtcd()

						//choose the new leader!
						leaderAddress := getAddressOfLowerKey()

						//check if I am the leader
						thisReplicaIsTheLeader = leaderAddress == (myIP + ":" + port)

						if thisReplicaIsTheLeader {
							myDebug("etcdElectionsWatcher: I AM THE NEW LEADER!!!!" + myIP + ":" + port)
							fmt.Println("I AM THE LEADER of view: ", leaderView)
						} else {
							myDebug("etcdElectionsWatcher: the new leader is: " + leaderAddress + ", view:" + strconv.Itoa(leaderView))
							fmt.Println("etcdElectionsWatcher: the new leader is: "+leaderAddress+", view:", strconv.Itoa(leaderView))
						}

						whenNewLeaderWasDiscovered = time.Now()
						sentMyVote = false

						//turn monitor on again
						if !leaderMonitorOn {
							//wait sometime to start the leaderMonitor again
							leaderMonitorOn = true

							go func() {
								//in fact, wait some seconds to really start the leader monitor
								//after a leader change, sometime is required to execute the requests
								time.Sleep(1 * time.Second)
								go leaderMonitor()
							}()
						}

						if thisReplicaIsTheLeader {

							//UPDATE the next available consensus instance!!!!
							mu.Lock()
							availableConsensusInstance = getLastExecutedConsensusInstance() + 1
							mu.Unlock()

							//check if there is another request with this consensus instance
							//TODO

							//if I am the leader, process the pending requests
							assignOrderToPendingRequests()
						}

					} else {
						//check if THIS REPLICA already have voted!
						nextView := strconv.Itoa(leaderView + 1)
						myVote := "VIEWCHANGE," + nextView + "," + myIP + ":" + port
						_, contains := electionMessages[myVote]
						if debug {
							mylog.Println("etcdelecwath: I will vote?: contains=", contains, ", sentMyVote=", sentMyVote)
						}
						if !contains && !sentMyVote {
							//VOTE!
							putViewChangeInEtcd(myVote)
							sentMyVote = true
							//IT COULD BE POSSIBLE TO VOTE TWICE: here, and triggered by the timer.
							//there is no problem, since the vote can enter only once in the map
							//however, it is an additional message in the etcd
						}
					}
				}
			}
		}()
	}
}

func getLastExecutedConsensusInstance() int {
	higher := 0
	for k := range valuesOfConsensusInstances {
		if k > higher {
			higher = k
		}
	}
	return higher
}

func assignOrderToPendingRequests() {
	for _, v := range requestsWaitingForTheLeader {
		//SOME REPEATED CODE
		req := mountRequest("0", v)
		mu.Lock()
		req.consensusInstance = availableConsensusInstance
		availableConsensusInstance++
		mu.Unlock()
		val := mountRequestValue(req)
		_, contains := inverseQueue[val]
		if !contains {
			putReqInEtcd(val)
		}
	}
	//empty list of requests Waiting!!
	requestsWaitingForTheLeader = nil
	requestsWaitingForTheLeader = make(map[string]string)
	requestsWaitingForTheLeaderTimes = nil
	requestsWaitingForTheLeaderTimes = make(map[string]time.Time)
}

func etcdQueueRemove(key string) {

	if !etcdOn && !connectEtcdClient() {
		mylog.Println("ERROR creating etcd client, see logs")
		return
	}
	myDelete := client.DeleteOptions{Recursive: true}
	_, err := kapi.Delete(context.Background(), key, &myDelete)
	if err != nil {
		mylog.Println("ERROR deleting ", key, " in etcd: ", err.Error())
	} else {
		mylog.Println("queue ", key, " deleted in etcd")
	}
}

func etcdWatcher(notify chan myRequest) {

	mylog.Println("watcher started")
	if !etcdOn && !connectEtcdClient() {
		mylog.Println("ERROR creating etcd client, see logs")
		return
	}
	myWatcher := client.WatcherOptions{AfterIndex: 0, Recursive: true}
	if watcherRestarts > 0 {
		nParts := strings.Split(lastObservedKey, "/")
		// format: /apprequests/0000000001234
		num := nParts[2]
		lastInd, _ := strconv.ParseUint(num, 10, 64)
		myWatcher = client.WatcherOptions{AfterIndex: lastInd, Recursive: true}
	}
	resp := kapi.Watcher(queueKey, &myWatcher)
	for {

		if debug {
			mylog.Println("watcher: waiting request from watcher...")
		}
		r, err := resp.Next(context.Background())
		if err != nil {
			mylog.Println("watcher: ERROR watching etcd: ", err.Error())

			//CHECK: SOMETIMES CLIENT CAN LOOSE CONNECTION WITH ETCD
			//IS IS NECESSARY TO RECONNECT
			mylog.Println("watcher: sending signal to monitor, and shutting down. Restarts until now: ", watcherRestarts, "; lastObservedKey: ", lastObservedKey)
			etcdOn = false
			watcherRestarts++
			restartWatcherChan <- "watcher"

			return
		}

		//process etcd notification

		go func() { //make the cycle of receiving notifications faster!! make it a Go routine!

			if r.Action == "create" {

				lastObservedKey = r.Node.Key // update last observed key by the watcher

				req := mountRequest(r.Node.Key, r.Node.Value) //mount a request

				//discover if this request cames from another replica
				/*
					requestFromOtherReplica := false
					if runningInKubernetes {
						requestFromOtherReplica = (req.containerIP != myIP)
					} else {
						requestFromOtherReplica = req.containerIP != (myIP + ":" + port)
					}
				*/
				myDebug("etcdWatcher: new request: " + r.Node.Key + " >> " + r.Node.Value)

				if loadingQueue {
					//save in a list to check after queue loading
					tmpRequestListWhileLoading[r.Node.Key] = r.Node.Value
				} else {

					//if requestFromOtherReplica && !executedRequest(req.consensusInstance) {
					if !executedRequest(req.consensusInstance) {

						if debug {
							mylog.Println("watcher: request was not executed yet: " + r.Node.Key + " >> " + r.Node.Value)
						}

						//is this a request WITHOUT DECIDED consensus order AND Am I the leader?
						if req.consensusInstance == 0 && thisReplicaIsTheLeader {

							//decide the order of this request!
							mu.Lock()
							req.consensusInstance = availableConsensusInstance
							availableConsensusInstance++
							mu.Unlock()

							if debug {
								mylog.Println("etcdWatc: setting order: ", req.consensusInstance, ", for ", req.requestID)
							}
						}

						if req.consensusInstance == 0 && !thisReplicaIsTheLeader {
							//add this request in a list that is waiting for the leader
							requestsWaitingForTheLeader[req.requestID] = r.Node.Value
							requestsWaitingForTheLeaderTimes[req.requestID] = time.Now()

							//start a timer to monitors the leader actuation
							if !leaderMonitorOn {
								go leaderMonitor()
								leaderMonitorOn = true
							}
						}

						//this requests has consensus instance?
						if req.consensusInstance != 0 {

							//SPECIAL SITUATION: a requests arrives, with consensus instance,
							//but IT WAS conflicting with another consensus that already exists
							//this can happens when a leader sets the CI, saves in etcd, but fails.
							//the next leader will receive this notification, and should update
							//this consensus instante: remove the old one and put the new!
							//this request did not achieved quorum, and stayed waiting, with
							//a consensus instance from a past view
							_, exists := requestIDofConsensusInstance[req.consensusInstance]
							if exists && thisReplicaIsTheLeader {
								if requestIDofConsensusInstance[req.consensusInstance] != req.requestID {

									if debug {
										mylog.Println("etcdWat: reassigning repeated CI")
									}
									//re-assign the consensus instance of this request!
									mu.Lock()
									req.consensusInstance = availableConsensusInstance
									availableConsensusInstance++
									mu.Unlock()

									//put in etcd
									val := createRequestValue(req.consensusInstance, req.quorumSize, req.containerIP, req.requestID, req.operation, req.view)
									putReqInEtcd(val)

									//remove the previous
									etcdQueueRemove(r.Node.Key)
								}
							}

							requestFromOtherReplica := false
							if runningInKubernetes {
								requestFromOtherReplica = (req.containerIP != myIP)
							} else {
								requestFromOtherReplica = req.containerIP != (myIP + ":" + port)
							}
							if requestFromOtherReplica {
								tmp := mountRequestValue(req)
								if !queueContainsKey(req.key) { //Is my queue mirrored with the etcd?
									queue[req.key] = tmp
									inverseQueue[tmp] = req.key
									insertReqInSpecificQueue(req.requestID, req.containerIP)
									if debug {
										mylog.Println("etcdWatcher: request from other replica, I did't had, added in lqueue: ", req.key, "/", req.requestID)
									}
								}
							}

							//Is it necessary send this to quorum checking?
							if !executedRequest(req.consensusInstance) {
								if debug {
									mylog.Println("etcdwatcher: notifying consensus watcher about: ", req.key, "-", req.requestID)
								}

								//if this requests was waiting for a leader assignment, we can undo that
								_, contains := requestsWaitingForTheLeader[req.requestID]
								if contains {
									delete(requestsWaitingForTheLeader, req.requestID)
									delete(requestsWaitingForTheLeaderTimes, req.requestID)
								}

								//create my registry of this request
								myRequestRegistryValue := createRequestValue(req.consensusInstance, req.quorumSize, myIP+":"+port, req.requestID, req.operation, req.view)

								// if this new registry is not in my queue,
								// i.e., if I already saved this requestID in etcd
								// and this notification is not about this previous saving
								_, contains2 := inverseQueue[myRequestRegistryValue]
								if !contains2 {
									if debug {
										mylog.Println("etcdwatch: saving in etcd: ", myRequestRegistryValue)
									}
									//save in etcd
									putReqInEtcd(myRequestRegistryValue)
								}

								go func() { //inform quorum observer about this new request
									notify <- req
								}()

							} else {
								if debug {
									mylog.Println("etcdWatch: not necessary to notify consensus about: ", req.requestID)
								}
							}
						} else {
							//there are $f$ remaining answers and the leader did not answered?
							//this is an indication of the need of a new leader
							//the leader executed a fault of timing!

							if debug {
								mylog.Print("etcdwatch: I have a message without consensus. ")
							}

							//some repeated action here

							qs := req.quorumSize
							majority := (qs / 2) + 1
							already, _ := countInQueueTurbo(req.requestID)
							if already >= majority {
								if debug {
									mylog.Println(" We already have majority, alt=", already, ", maj=", majority, ", we HAVE TO CHANGE LEADER!")
								}
								//start elections!
								//startElections()
							} else {
								if debug {
									mylog.Println(" alt=", already, ", maj=", majority, ", we can wait for a leader answer")
								}
							}
						}
					} else {
						if debug {
							mylog.Println("watcher: request from THIS replica (do nothing), ", req.requestID, ":", req.containerIP, "; or already executed:", executedRequest(req.consensusInstance))
						}
					}
					//go func moved from here. You do not have to notify yourself
					//if this watcher was from your own receiving of request
				}
			}
		}()
	}
}

func watcherMonitor(ch chan string) {
	var msg string
	for {
		msg = <-ch //wait for restart messages
		if msg == "watcher" {
			mylog.Println("watcherMonit: restarting watcher goroutine...")
			go etcdWatcher(requestChan) //restart etcd watcher!
		}
	}
}

func executeRequest(value int, operation string) int {
	if operation == "inc" {
		value = value + 1
	} else if operation == "dou" {
		if value > 30 {
			value = value / 2
		}
	}
	return value
}

//200 execuções, 3 réplicas: 2,8% de memória alocada permanente
