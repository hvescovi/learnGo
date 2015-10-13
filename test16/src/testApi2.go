package main

import (
  "fmt"
  "github.com/parnurzeal/gorequest"
)


func main() {
	 request := gorequest.New()
   _, body, _ := request.Get("http://127.0.0.1:4001/v2/keys/foo").End()
   fmt.Println(body)

   //IT WORKS!!
   save := gorequest.New()
   _, body2, _ := save.Put("http://127.0.0.1:4001/v2/keys/foo?value=hi").End()
   fmt.Println(body2)

  request = gorequest.New()
  _, body, _ = request.Get("http://127.0.0.1:4001/v2/keys/foo").End()
  fmt.Println(body)
 }
