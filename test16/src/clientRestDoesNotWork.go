package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	//"strconv"
)

func main() {
	apiUrl := "http://192.168.15.100:8080"
	resource := "/"

	data := url.Values{}
	data.Set("request", "lalala")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	//r.Header.Add("Authorization", "auth_token=\"XXXXXX\"")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(r)
	fmt.Println(resp.Status)

	content, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(content)
}
