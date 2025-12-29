package serverStub

import (
	"awesomeproject/handler"
	"net/rpc"
)

// 鸭子模型： 实现了Hello 方法的Struct 都可以堪称 HelloServicer
type HelloServicer interface {
	Hello(request string, reply *string) error
}

func RegisterHelloService(service HelloServicer) error {
	err := rpc.RegisterName(handler.HelloServiceName, service)
	return err
}
