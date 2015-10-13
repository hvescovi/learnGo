package main

import (
  "fmt"
  "log"
  "net/http"
  "encoding/json"

  "github.com/gorilla/mux"
)

type Movie struct{
  Title string `json:"title"`
  Rating string `json:"rating"`
  Year string `json:"year"`
}

var movies = map[string]*Movie{
  "tt001": &Movie{Title:"Star Wars: a new hope", Rating:"8.5", Year:"1977"},
  "tt002": &Movie{Title:"Indiana Jones and the last cruzade", Rating:"9", Year:"1985"},
}

func main() {
  router := mux.NewRouter()
  router.HandleFunc("/movies", handleMovies).Methods("GET")
  router.HandleFunc("/movies/{key}", handleMovie).Methods("GET")
  http.ListenAndServe(":8080", router)
}

func handleMovie(res http.ResponseWriter, req *http.Request) {
  res.Header().Set("Content-Type", "application/json")

  vars:= mux.Vars(req)
  key := vars["key"]

  if movie, ok := movies[key]; ok {
    outgoingJSON, err := json.Marshal(movie)
    if err != nil {
      log.Println(err.Error())
      http.Error(res, err.Error(), http.StatusInternalServerError)
      return
    }
    fmt.Fprint(res, string(outgoingJSON))
  } else {
    res.WriteHeader(http.StatusNotFound)
    fmt.Fprint(res, string("Movie not found"))
  }
}

func handleMovies(res http.ResponseWriter, req *http.Request) {
  res.Header().Set("Content-Type", "application/json")

  outgoingJSON, err := json.Marshal(movies)

  if err != nil {
    log.Println(err.Error())
    http.Error(res, err.Error(), http.StatusInternalServerError)
    return
  }

  fmt.Fprint(res, string(outgoingJSON))
}
