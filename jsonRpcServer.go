package main

import (
	"net"
	"net/rpc/jsonrpc"
)
import "net/rpc"

type HelloService2 struct{}

func (s *HelloService2) Hello(request string, reply *string) error {
	*reply = "Hello " + request
	return nil
}

func main() {
	_ = rpc.RegisterName("HelloService", &HelloService2{})
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic("接收")
		}
		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
