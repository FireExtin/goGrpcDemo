package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	if err := client.Call("HelloService.Hello", "Xc", &reply); err != nil {
		log.Fatal("call failed:", err)
	}
	fmt.Println(reply)
}
