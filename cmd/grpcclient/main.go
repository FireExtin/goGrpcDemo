package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "awesomeproject/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 定义命令行参数
	addr := flag.String("addr", "127.0.0.1:50051", "server address")
	name := flag.String("name", "world", "name to greet")
	flag.Parse()

	// 【优化】使用 grpc.NewClient 替换已弃用的 grpc.Dial
	// NewClient 是现代 gRPC-Go 推荐的连接方式
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("could not close connection: %v", err)
		}
	}(conn) // 确保程序退出时关闭连接

	// 创建服务客户端
	client := pb.NewGreeterClient(conn)

	// 设置超时上下文，防止请求长时间阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 调用 SayHello
	hello, err := client.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", hello.GetMessage())

	// 调用 SayBye
	bye, err := client.SayBye(ctx, &pb.ByeRequest{Name: *name, Message: "Goodbye"})
	if err != nil {
		log.Fatalf("could not say bye: %v", err)
	}
	log.Printf("Farewell: %s", bye.GetMessage())
}
