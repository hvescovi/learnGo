package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	resp, err := http.PostForm("http://192.168.15.100:7500", url.Values{"request": {"reqB"}})
	if err != nil {
		fmt.Println("error creating the post request: ", err)
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("error reading data", err)
		} else {
			fmt.Println(string(body))
		}
	}
}
