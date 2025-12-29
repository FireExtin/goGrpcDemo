package hertzrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Standard JSON-RPC 2.0 Request
type Request struct {
	JsonRpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	Id      interface{}       `json:"id"`
}

// Standard JSON-RPC 2.0 Response
type Response struct {
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	Id      interface{} `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

type Dispatcher struct {
	mu         sync.RWMutex
	serviceMap map[string]*service
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		serviceMap: make(map[string]*service),
	}
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the name.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// RegisterName registers the service with the given name.
// It mimics net/rpc's registration logic.
func (d *Dispatcher) RegisterName(name string, rcvr interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = name
	s.method = make(map[string]*methodType)

	// Install the methods
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)
		mtype := method.Type
		mname := method.Name

		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			continue
		}
		// Method needs one out: error.
		if mtype.NumOut() != 1 {
			continue
		}
		if returnType := mtype.Out(0); returnType != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		s.method[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}

	if len(s.method) == 0 {
		return fmt.Errorf("hertzrpc: %q has no exported methods of suitable type", s.name)
	}
	d.serviceMap[s.name] = s
	return nil
}

// Handle processes the Hertz request as a JSON-RPC request
func (d *Dispatcher) Handle(ctx context.Context, c *app.RequestContext) {
	var req Request
	// Use BindAndValidate or just Bind. Bind checks Content-Type.
	if err := c.Bind(&req); err != nil {
		c.JSON(consts.StatusOK, Response{
			JsonRpc: "2.0",
			Error:   &Error{Code: -32700, Message: "Parse error: " + err.Error()},
			Id:      nil,
		})
		return
	}

	// 1. Parse Method Name (Service.Method)
	dot := strings.LastIndex(req.Method, ".")
	if dot < 0 {
		d.sendError(c, req.Id, -32600, "Invalid Request: Method name format Service.Method")
		return
	}
	serviceName := req.Method[:dot]
	methodName := req.Method[dot+1:]

	// 2. Look up Service
	d.mu.RLock()
	svc, ok := d.serviceMap[serviceName]
	d.mu.RUnlock()
	if !ok {
		d.sendError(c, req.Id, -32601, fmt.Sprintf("Service not found: %s", serviceName))
		return
	}

	// 3. Look up Method
	mtype, ok := svc.method[methodName]
	if !ok {
		d.sendError(c, req.Id, -32601, fmt.Sprintf("Method not found: %s", methodName))
		return
	}

	// 4. Parse Args
	var argv reflect.Value
	argIsValue := false // if true, need to indirect before calling.

	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}

	// argument unmarshaling
	if len(req.Params) > 0 {
		// We assume params[0] corresponds to the first argument
		// (This supports positional params with length 1, which matches standard Go RPC)
		if err := json.Unmarshal(req.Params[0], argv.Interface()); err != nil {
			d.sendError(c, req.Id, -32602, "Invalid params: "+err.Error())
			return
		}
	}

	if argIsValue {
		argv = argv.Elem()
	}

	// 5. Prepare Reply
	replyv := reflect.New(mtype.ReplyType.Elem())

	// 6. Call
	function := mtype.method.Func
	// Func.Call expects [Receiver, Arg, Reply]
	returnValues := function.Call([]reflect.Value{svc.rcvr, argv, replyv})

	// 7. Check Error
	errInter := returnValues[0].Interface()
	if errInter != nil {
		d.sendError(c, req.Id, -32000, errInter.(error).Error())
		return
	}

	// 8. Success
	c.JSON(consts.StatusOK, Response{
		JsonRpc: "2.0",
		Result:  replyv.Interface(),
		Id:      req.Id,
	})
}

func (d *Dispatcher) sendError(c *app.RequestContext, id interface{}, code int, msg string) {
	c.JSON(consts.StatusOK, Response{
		JsonRpc: "2.0",
		Error:   &Error{Code: code, Message: msg},
		Id:      id,
	})
}
