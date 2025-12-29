package main

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"awesomeproject/handler"
	"awesomeproject/serverStub"
)

func main() {
	err := serverStub.RegisterHelloService(&handler.HelloService{})
	if err != nil {
		panic("failed to register hello service: " + err.Error())
	}
	http.HandleFunc("/jsonRpc", func(w http.ResponseWriter, r *http.Request) {
		var conn io.ReadWriteCloser = struct {
			io.Writer
			io.ReadCloser
		}{
			ReadCloser: r.Body,
			Writer:     w,
		}
		err := rpc.ServeRequest(jsonrpc.NewServerCodec(conn))
		if err != nil {
			return
		}
	})
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
}
