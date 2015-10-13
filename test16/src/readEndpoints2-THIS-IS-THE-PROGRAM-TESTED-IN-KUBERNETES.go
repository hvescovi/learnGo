package main

import (
  "net/http"
  "fmt"
  "os"
  "io/ioutil"
  "strings"
)

func main() {

  resp, err := http.Get("http://192.168.15.100:8080/api/v1/endpoints")
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  } 
  defer resp.Body.Close()
  contentByte, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  
  content := string(contentByte)
  
  words := strings.Split(content,"\"ip\":")
  for _,v := range words {
    fmt.Println("checking: ",v[1:3])
    if v[1:3] == "18" { // 18.x.x.x.x
      fmt.Println("=======> IP: ", v[1:11])
    } else {
    //fmt.Println(k, "=", v)
    }
  }
  fmt.Println("ok, unmarshalled")
}
