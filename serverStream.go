package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"awesomeproject/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata" // 导入 metadata 包
)

const ADDR = ":50052"

// --- 魔法 1: 拦截器 (Interceptors) ---

// UnaryInterceptor: 处理普通非流式请求的“管家”
func myUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("【拦截器】收到普通请求: %s", info.FullMethod)
	// 在这里可以统一检查 Token
	return handler(ctx, req)
}

type wrappedStream struct {
	grpc.ServerStream
	count int
}

func (w *wrappedStream) SendMsg(m any) error {
	w.count++
	log.Printf("Filtering: processing %d messages", w.count)
	return w.ServerStream.SendMsg(m)
}

// StreamInterceptor: 处理流式请求的“管家”
func myStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Printf("【拦截器】收到流式请求: %s", info.FullMethod)
	wrapper := &wrappedStream{ServerStream: ss}
	// 继续执行后续逻辑
	err := handler(srv, wrapper)
	log.Printf("[Filtering end] %d messages processed", wrapper.count)
	return err
}

type serverStream struct {
	proto.UnimplementedGreeterServer
}

func (s *serverStream) AllStream(allStr proto.Greeter_AllStreamServer) error {
	// --- 魔法 2: 提取 Metadata (小纸条) ---
	// 从流的 Context 中获取客户端传来的元数据
	md, ok := metadata.FromIncomingContext(allStr.Context())
	if ok {
		fmt.Printf("【服务端】收到小纸条 Metadata: %v\n", md)
		if tokens, exists := md["token"]; exists {
			fmt.Printf("【服务端】验证 Token: %s\n", tokens[0])
		}
	}

	// 保留原有的 Context 逻辑
	ctx, cancel := context.WithTimeout(allStr.Context(), 20*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				data, err := allStr.Recv()
				if err == io.EOF {
					return
				}
				if err != nil {
					return
				}
				fmt.Println("收到客户端消息: " + data.Data)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 2):
				_ = allStr.Send(&proto.StreamResData{Data: "服务端心跳包"})
			}
		}
	}()

	wg.Wait()
	return nil
}

// 实现必要的接口
func (s *serverStream) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloReply, error) {
	return &proto.HelloReply{Message: "Hello " + req.GetName()}, nil
}

func main() {
	listen, _ := net.Listen("tcp", ADDR)

	// 在创建服务器时，注册拦截器
	s := grpc.NewServer(
		grpc.UnaryInterceptor(myUnaryInterceptor),   // 注册普通拦截器
		grpc.StreamInterceptor(myStreamInterceptor), // 注册流式拦截器
	)

	proto.RegisterGreeterServer(s, &serverStream{})
	log.Printf("流式服务(带拦截器)已启动: %s", ADDR)
	s.Serve(listen)
}
