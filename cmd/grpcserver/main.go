package main

import (
	"context"
	"flag"
	"log"
	"net"

	pb "awesomeproject/proto"
	"google.golang.org/grpc"
)

type greeterServer struct {
	pb.UnimplementedGreeterServer
}

func (s *greeterServer) SayHello(_ context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + req.GetName()}, nil
}

func (s *greeterServer) SayBye(_ context.Context, req *pb.ByeRequest) (*pb.HelloReply, error) {
	message := req.GetMessage()
	if message == "" {
		message = "Bye"
	}
	return &pb.HelloReply{Message: message + " " + req.GetName()}, nil
}

func main() {
	addr := flag.String("addr", "127.0.0.1:50051", "listen address")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}

	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &greeterServer{})

	log.Printf("gRPC server listening on %s", *addr)
	log.Fatal(server.Serve(listener))
}
