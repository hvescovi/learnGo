package main

import (
  "net/http"
  "fmt"
  "os"
  "io/ioutil"
  "strings"
)

func main() {

  fmt.Println(len(os.Args))
  if len(os.Args) < 2 {
    fmt.Println("please, specify the value to set in foo")
    os.Exit(1)
  }

  value := os.Args[1]
  fmt.Println("value: ",value)

  response, err := http.Post("http://127.0.0.1:4001/v2/keys/", "text", strings.NewReader("hello"))
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  } else {
    defer response.Body.Close()
    contentByte, err := ioutil.ReadAll(response.Body)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    content := string(contentByte)
    fmt.Println("response: ",content)
  }
}
