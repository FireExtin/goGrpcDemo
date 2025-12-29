package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	err = client.Call("HelloService.Hello", "Xc", &reply)
	if err != nil {
		log.Fatal("call failed:", err)
	}
	fmt.Println(reply)
}
