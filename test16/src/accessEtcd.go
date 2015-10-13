package main

import (
    "log"
    "time"
    "fmt"

    "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
    "github.com/coreos/etcd/client"


    //"fmt"
)

func main() {
    cfg := client.Config{
        Endpoints:               []string{"http://127.0.0.1:2379"},
        Transport:               client.DefaultTransport,
        // set timeout per request to fail fast when the target endpoint is unavailable
        HeaderTimeoutPerRequest: time.Second,
    }
    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)

    //resp, err := kapi.Set(context.Background(), "foo", "bar", nil)
    resp, err := kapi.Get(context.Background(), "foo", nil)
    fmt.Println(resp.Node.Value)
    if err != nil {
        log.Fatal(err)
    }
    // else {
    //    log.Print(resp)
    //}
/*
    f := getAction {Key: "/foo/",Recursive: false,Sorted: false, Quorum: false}
    ep:= url.URL{Scheme: "http", Host: "localhost", Path: "/"}
    wantHeader := http.Header{}
    baseWantURL := &url.URL{Scheme: "http", Host: "localhost", Path: "/foo"}
    got := *f.HTTPRequest(ep)
    wantURL := baseWantURL
    wantURL.RawQuery = "quorum=false&recursive=false&sorted=false"
    err2 := assertRequest(got, "GET", wantURL, wantHeader, nil)
    if err2 != nil {
      fmt.Println("#%d: %v", i, err2)
    } else {
      fmt.Println(" test ok")
    }
    */
}
