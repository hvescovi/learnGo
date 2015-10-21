package main

//CLIENTS CAN EXECUTE LIKE:

// curl localhost:7500
// OR more exciting:
// ab -n 1000 -c 10 http://127.0.0.1:7500/

import (
	"net/http"
	"fmt"
	"sync"
	"io/ioutil"
)

const DEBUG = true

func main() {

    var wg sync.WaitGroup
    wg.Add(2) //block the finish of the program while 2 thread alives

    //listen requests directly on port 7500
    go listen()
	
	fmt.Println("server started")
	
	wg.Wait()
}

func listen() {
	
	//listen requests
	http.HandleFunc("/", handlerListenDirectly)
	http.HandleFunc("/mid", handlerListenViaMiddleware)
	fmt.Println("listening on port 7500")
	http.ListenAndServe(":7500", nil)
}

func handlerListenDirectly(w http.ResponseWriter, r *http.Request) {

    //fmt.Print("request incoming; ")
	
	//read the command
	body, _ := ioutil.ReadAll(r.Body)
    
    //fmt.Print("command received: ", body, ";")
	
	fmt.Fprintf(w, "ok, your command: ")
	fmt.Fprintf(w, string(body))
	fmt.Fprintf(w, " was executed")
	//fmt.Println("request executed")
	fmt.Print(".")
}

func handlerListenViaMiddleware(w http.ResponseWriter, r *http.Request) {
	
	//fmt.Print("[listened via middleware, fowarded]")
	//fmt.Print("f")
	//fmt.Fprintf(w, "ok, listened via middlware.\n")
	//fmt.Println("request fowarded")
	http.Redirect(w, r, "/", 301)
}
