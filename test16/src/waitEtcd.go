package main

//
// A GOOD EXAMPLE!!!!!!!
// https://gist.github.com/tcotav/7df72a406a2f82b6357d
//
//

import (
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

func main() {
	cfg := client.Config{
		Endpoints: []string{"http://192.168.15.100:4001"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	//resp, err := kapi.Set(context.Background(), "foo", "bar", nil)
	//var ops *WatcherOptions
	myOps := client.WatcherOptions{AfterIndex: 0, Recursive: true}
	resp := kapi.Watcher("/apprequests", &myOps)
	for {
		r, err := resp.Next(context.Background())
		if err != nil {
			fmt.Println("error!", err)
		}
		action := r.Action
		fmt.Println("action: ", action)
	}

	//if err != nil {
	//      log.Fatal(err)
	//  }
}
