package main

import (
	"log"
	"net/http"
	"net/rpc"

	"awesomeproject/handler"
	"awesomeproject/serverStub"
)

func main() {
	if err := serverStub.RegisterHelloService(&handler.HelloService{}); err != nil {
		log.Fatal("failed to register hello service:", err)
	}

	rpc.HandleHTTP()
	addr := ":8081"
	log.Printf("net/rpc over HTTP listening on %s (path: %s)", addr, rpc.DefaultRPCPath)
	log.Fatal(http.ListenAndServe(addr, nil))
}
