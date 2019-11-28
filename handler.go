package socketio

import (
	"fmt"
	"reflect"
)

// Context is mainly struct in socketio library
type Context struct {
	Namespace string //
	Event     string //
	Message   string //

	Connect *Conn //

	Errors error //TODO: migrate from gin/gonic

	argTypes []reflect.Type
	f        reflect.Value
}

func (c *Context) Error(err error) {

}

func newEventFunc(f ...HandlerFunc) {
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		panic("event handler must be a func.")
	}
	ft := fv.Type()
	if ft.NumIn() < 1 || ft.In(0).Name() != "Conn" {
		panic("handler function should be like func(socketio.Conn, ...)")
	}
	argTypes := make([]reflect.Type, ft.NumIn()-1)
	for i := range argTypes {
		argTypes[i] = ft.In(i + 1)
	}
	if len(argTypes) == 0 {
		argTypes = nil
	}

	// return &Context{
	// 	argTypes: argTypes,
	// 	f:        fv,
	// }
}

func newAckFunc(f interface{}) *Context {
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		panic("ack callback must be a func.")
	}
	ft := fv.Type()
	argTypes := make([]reflect.Type, ft.NumIn())
	for i := range argTypes {
		argTypes[i] = ft.In(i)
	}
	if len(argTypes) == 0 {
		argTypes = nil
	}
	return &Context{
		argTypes: argTypes,
		f:        fv,
	}
}

// Call
func (h *Context) Call(args []reflect.Value) (ret []reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("event call error: %s", r)
			}
		}
	}()
	ret = h.f.Call(args)
	return
}
