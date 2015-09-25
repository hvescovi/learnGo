package main

import "fmt"

func main() {
  done := make(chan bool)
    
  go func() {
    fmt.Println("Hello everybody")
    done <- true
  }()

  fmt.Println("Bye people")
  <-done
}
