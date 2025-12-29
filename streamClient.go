package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"awesomeproject/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata" // 导入元数据包
)

// --- 魔法 1: 客户端拦截器 ---

// clientUnaryInterceptor: 记录普通请求耗时
func clientUnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("【客户端拦截器】普通请求 %s 耗时: %v", method, time.Since(start))
	return err
}

// clientStreamInterceptor: 记录流式请求开启
func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Printf("【客户端拦截器】开启流式请求: %s", method)
	return streamer(ctx, desc, cc, method, opts...)
}

func main() {
	// 1. 建立连接时注册拦截器
	conn, err := grpc.NewClient(
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientUnaryInterceptor),   // 注册普通拦截器
		grpc.WithChainStreamInterceptor(clientStreamInterceptor), // 注册流式拦截器
	)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := proto.NewGreeterClient(conn)

	// ==========================================
	// 2. 准备 Metadata (小纸条)
	// ==========================================
	md := metadata.Pairs("token", "my-secret-token", "client-version", "1.0.0")

	// ==========================================
	// 3. 结合 Context 取消逻辑
	// ==========================================
	fmt.Println("\n--- 开始测试: 双向流 (Context取消 + Metadata + 拦截器) ---")

	// 创建可取消的 Context，并把 Metadata 注入进去
	// 注意：NewOutgoingContext 接收一个父 Context，我们把 Background 传进去
	baseCtx, cancel := context.WithCancel(context.Background())
	ctx := metadata.NewOutgoingContext(baseCtx, md)

	allStream, err := c.AllStream(ctx)
	if err != nil {
		log.Fatalf("开启流失败: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	// 协程 1: 持续接收 (保持原逻辑)
	go func() {
		defer wg.Done()
		for {
			resp, err := allStream.Recv()
			if err == io.EOF {
				fmt.Println("[客户端] 服务端正常关闭")
				return
			}
			if err != nil {
				fmt.Printf("[客户端] 接收停止 (原因: %v)\n", err)
				return
			}
			fmt.Printf("[客户端] 收到: %s\n", resp.Data)
		}
	}()

	// 协程 2: 模拟运行并主动取消 (保持原逻辑)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			err := allStream.Send(&proto.StreamReqData{Data: fmt.Sprintf("消息 %d", i)})
			if err != nil {
				return
			}
			time.Sleep(time.Second)

			if i == 4 {
				fmt.Println("\n[客户端] !!! 触发主动取消 !!!")
				cancel() // 调用 cancel 会让 ctx 结束，从而让流关闭
				return
			}
		}
	}()

	wg.Wait()
	fmt.Println("测试结束。")
}
