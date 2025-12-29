package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"awesomeproject/handler"
	"awesomeproject/pkg/hertzrpc"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/websocket"
)

var upgrader = websocket.HertzUpgrader{
	CheckOrigin: func(ctx *app.RequestContext) bool {
		return true
	},
}

func main() {
	// 1. Initialize Hertz Server
	h := server.Default(server.WithHostPorts(":8082"))

	// 2. Initialize JSON-RPC Dispatcher
	// This replaces the manual switch-case logic with a reflection-based dispatcher
	dispatcher := hertzrpc.NewDispatcher()
	
	// Register the service (just like net/rpc)
	// This solves the scalability issue: you can register as many services as you want
	if err := dispatcher.RegisterName("HelloService", new(handler.HelloService)); err != nil {
		log.Fatal("Failed to register HelloService:", err)
	}

	// 3. Register JSON-RPC Route
	// All JSON-RPC requests go here
	h.POST("/jsonRpc", dispatcher.Handle)

	// 4. Register WebSocket Route (Streaming)
	// This handles the streaming requirement natively in Hertz
	h.GET("/stream", func(c context.Context, ctx *app.RequestContext) {
		if err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
			handleStream(conn)
		}); err != nil {
			log.Print("upgrade:", err)
			return
		}
	})

	fmt.Println("Hertz server is running on :8082")
	fmt.Println(" - JSON-RPC: http://127.0.0.1:8082/jsonRpc")
	fmt.Println(" - Streaming: ws://127.0.0.1:8082/stream")
	h.Spin()
}

// handleStream mimics the bidirectional streaming logic from your gRPC example
func handleStream(conn *websocket.Conn) {
	defer conn.Close()

	// Channel to signal completion
	done := make(chan struct{})

	// Goroutine 1: Read messages from client (Recv)
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("[Server] Read error:", err)
				return
			}
			log.Printf("[Server] Received from client: %s", message)
		}
	}()

	// Goroutine 2: Send messages to client (Send)
	// Simulating server push / heartbeat
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			return
		default:
			msg := fmt.Sprintf("Server Heartbeat %d", i)
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("[Server] Write error:", err)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}

	// Wait for the client to close the connection or an error to occur
	<-done
}
