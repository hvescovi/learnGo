package main

//curl http://127.0.0.1:4001/v2/keys/queue/ -XPOST -d value=pedido1
//{"action":"create","node":{"key":"/queue/00000000000000000037","value":"pedido1","modifiedIndex":37,"createdIndex":37}}

import (
	"fmt"
	"math/rand"
	"os"
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
		fmt.Println(err)
		os.Exit(1)
	}
	kapi := client.NewKeysAPI(c)

	//setting a value which key will be automaticaly incremented

	value := RandStringRunes(5)

	//resp, err := kapi.Set(context.Background(), "foo", value, nil)
	//resp, err := kapi.Create(context.Background(), "/myrequests", value)
	resp, err := kapi.CreateInOrder(context.Background(), "/myrequests2", value, nil)

	//resp, err := kapi.Get(context.Background(), "foo", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(resp.Node.Value)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
