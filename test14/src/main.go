package main

import (
  "fmt"
  "log"
  "net/http"

  "github.com/gorilla/mux"
)

func main() {
  router := mux.NewRouter().StrictSlash(true)
  router.HandleFunc("/hello/{name}", index).Methods("GET")
  log.Fatal(http.ListenAndServe(":8080", router))
}

func index(w http.ResponseWriter, r *http.Request) {
  log.Println("Responsing to /hello request\n",r.UserAgent())

  vars := mux.Vars(r)
  name := vars["name"]

  w.WriteHeader(http.StatusOK)
  fmt.Fprintln(w, "Hello: ", name)
}
