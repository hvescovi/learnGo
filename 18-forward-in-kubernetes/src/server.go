package main

//CLIENTS CAN EXECUTE LIKE:

// curl localhost:7500
// OR more exciting:
// ab -n 1000 -c 10 http://127.0.0.1:7500/

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"myutils"
	"os"
	"log"
	"net/url"
	"strings"
	"html"
)

const DEBUG = true

var (
	leaderIP string
	logFilename string
	health string = ""
	port string
)

func main() {
	myutils.RandInit()
	logFilename = "/tmp/replica-log-" + myutils.RandStringRunes(5)
	logFile, err := os.Create(logFilename)
	if err != nil {
		health += " could not create log file: " + logFilename + ": " + err.Error()
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	leaderIP = ""

  //exists port? localhost running!
	if len(os.Args) > 0 {
		port = os.Args[1]
	} else {
		port = "8000"
	}

	//listen requests
	http.HandleFunc("/command", handlerCommand)
	http.HandleFunc("/log",handlerLog)
	http.HandleFunc("/healthz", handlerHealthz)
	http.HandleFunc("/", handlerLeader)
	//http.HandleFunc("/", handlerDefault)
	fmt.Println("listening on port ",port)
	http.ListenAndServe(":"+port, nil)
}

func handlerLog(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "displaying log: \n")
	content, err := ioutil.ReadFile(logFilename)
	if err != nil {
		fmt.Fprintf(w, " could not print the log, " + err.Error())
	} else {
		fmt.Fprintf(w, string(content))
	}
}

func handlerCommand(w http.ResponseWriter, r *http.Request) {
	//read the command
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Fprintf(w, "thank you for your calling: ")
	fmt.Fprintf(w, string(body))
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "healthz \n")
	fmt.Fprintf(w, health)
}

func handlerDefault(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "default action \n")
}

func handlerLeader(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	fmt.Fprintf(w, "trying to set leader \n")
	s := r.URL.Path
	fmt.Fprintf(w, "\n\n",s," \n\n")
	x := html.EscapeString(r.URL.Path)
	fmt.Fprintf(w, "\n x => " , x, " \n")
	parts1 := strings.Split(s, "%")
	for k1,v1 := range parts1 {
		fmt.Fprintf(w, "\n k=",k1,",v=",v1)
	}
	//fmt.Fprintf(w, parts[1]," \n")

	u, err := url.Parse(s)
	if err != nil {
		fmt.Fprintf(w, "error receiving parameters, " + err.Error())
	} else {
		//fmt.Fprintf(w, "raw: ",u.RawQuery)
		parts := strings.Split(u.Path, "/")
		for k,v := range parts {
			fmt.Fprintf(w, "\n k=",k,",v=",v)
		}
		/*
	  if len(parts) != 2 {
			fmt.Fprintf(w, "usage: setleader/ip-of-the-leader")
		} else {
			leaderIP := parts[1]
			fmt.Fprintf(w, "leader = ",leaderIP)
		}
		*/
	}
}

func handlerRedirect(w http.ResponseWriter, r *http.Request) {
	s := r.URL.String()
	s = leaderIP + "/" + s
	log.Println("redirecting! new command: " + s)
	http.Redirect(w, r, s, 301)
}
