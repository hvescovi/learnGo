package main

import (
  "strconv"
  "io/ioutil"
)

var idx int

func main() {
  idx = 1
  saveMessage("ola pessoal")
  saveMessage("tudo bem")
  saveMessage("como vao")
}
func saveMessage(msg string) {
	data := []byte(msg)
	fileName := "/tmp/replicalog-" + strconv.Itoa(idx)
	idx++
	ioutil.WriteFile(fileName, data, 0644)
}
