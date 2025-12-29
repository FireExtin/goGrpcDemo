package handler

const HelloServiceName = "HelloService"

type HelloService struct{}

func (h *HelloService) Hello(request string, reply *string) error {
	*reply = "Hello " + request + " from ppp"
	return nil
}
