package main

import (
  "net/http"
  "fmt"
  "os"
  "io/ioutil"
)

func main() {

  resp, err := http.Get("http://192.168.15.100:8080/api/v1/endpoints")
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  } 
  defer resp.Body.Close()
  content, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  fmt.Println(string(content))
}
