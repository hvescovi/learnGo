package main

import (
  "net/http"
  "fmt"
  "os"
  "bytes"
  "strconv"
  "net/http/httputil"
  "net/url"

)

func main() {

  fmt.Println(len(os.Args))
  if len(os.Args) < 2 {
    fmt.Println("please, specify the value to set in foo")
    os.Exit(1)
  }

  value := os.Args[1]
  fmt.Println("value: ",value)

  put := url.Values{}
  put.Set("value", value)
  put.Add("ttl","")
  encode := put.Encode()
  req, err := http.NewRequest("PUT","http://127.0.0.1:4001/v2/keys/foo", bytes.NewBufferString(encode))
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  req.Header.Add("Content-Type","application/x-www-form-urlencoded")
  //req.Header.Add("Content-Length",strconv.Itoa(len(encode)))
  req.Header.Add("X-Content-Length", strconv.Itoa(len(encode)))
  dump, err := httputil.DumpRequestOut(req, true)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  fmt.Println(string(dump))
}
