package main

import "net"
import "net/rpc"

type HelloService struct{}

func (s *HelloService) Hello(request string, reply *string) error {
	*reply = "Hello " + request
	return nil
}

func main() {
	_ = rpc.RegisterName("HelloService", &HelloService{})
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	conn, err := listener.Accept()
	if err != nil {
		panic("failed to accept: " + err.Error())
	}
	rpc.ServeConn(conn)
}
