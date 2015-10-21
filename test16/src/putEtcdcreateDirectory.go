package main

import (
    "log"
    "time"
    "fmt"

    "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
    "github.com/coreos/etcd/client"
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
 
    //myOps := client.SetOptions{Dir: true}
    //resp, err := kapi.Set(context.Background(), "/listnames/", "", &myOps)
	resp, err := kapi.Set(context.Background(), "/listnames/1", "18.16.32.16,18.16.51.2,R0C0", nil)
    //resp, err := kapi.Get(context.Background(), "foo", nil)
    fmt.Println(resp.Node.Value)
    if err != nil {
        log.Fatal(err)
    }
}
