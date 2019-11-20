package socketio

import (
	"reflect"
	"testing"

	"github.com/googollee/go-socket.io/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// type TestCase struct {
// }

// type ErrorTestCase struct {
// 	Message string
// }

// type TestCase struct {
// }

// type ErrorTestCase struct {
// 	Message string
// }

func TestNamespaceHandler(t *testing.T) {
	should := assert.New(t)
	must := require.New(t)
	h := newHandler()

	onConnectCalled := false
	h.OnConnect(func(c Conn) error {
		onConnectCalled = true
		return nil
	})

	disconnectMsg := ""
	h.OnDisconnect(func(c Conn, msg string) {
		disconnectMsg = msg
	})

	var onerror error
	h.OnError(func(err error) {
		onerror = err
	})

	header := parser.Header{}

	header.Type = parser.Connect
	args := h.getTypes(header, "")
	should.Nil(args)
	h.dispatch(&namespaceConn{}, header, "", nil)
	should.True(onConnectCalled)

	header.Type = parser.Disconnect
	args = h.getTypes(header, "")
	should.Equal([]reflect.Type{reflect.TypeOf("")}, args)
	h.dispatch(&namespaceConn{}, header, "", []reflect.Value{reflect.ValueOf("disconn")})
	should.Equal("disconn", disconnectMsg)

	header.Type = parser.Error
	args = h.getTypes(header, "")
	should.Equal([]reflect.Type{reflect.TypeOf("")}, args)
	h.dispatch(&namespaceConn{}, header, "", []reflect.Value{reflect.ValueOf("failed")})
	should.Equal(onerror.Error(), "failed")

	header.Type = parser.Event
	args = h.getTypes(header, "nonexist")
	should.Nil(args)
	ret, err := h.dispatch(&namespaceConn{}, header, "nonexist", nil)
	must.Nil(err)
	should.Nil(ret)
}

func TestNamespaceHandlerEvent(t *testing.T) {
	tests := []struct {
		name string

		events   []string
		handlers []interface{}

		event string
		args  []interface{}

		ok  bool
		ret []interface{}
	}{
		{
			name: "string handler",

			events: []string{"e", "n"},
			handlers: []interface{}{
				func(c Conn, str string) string {
					return "handled " + str
				},
				func(c Conn) {},
			},

			event: "e",
			args:  []interface{}{"str"},

			ok:  true,
			ret: []interface{}{"handled str"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			should := assert.New(t)
			must := require.New(t)

			h := newHandler()
			for i, e := range test.events {
				h.OnEvent(e, test.handlers[i])
			}

			header := parser.Header{
				Type: parser.Event,
			}
			target := make([]reflect.Type, len(test.args))
			args := make([]reflect.Value, len(test.args))
			for i := range test.args {
				target[i] = reflect.TypeOf(test.args[i])
				args[i] = reflect.ValueOf(test.args[i])
			}
			types := h.getTypes(header, test.event)
			should.Equal(target, types)
			ret, err := h.dispatch(&namespaceConn{}, header, test.event, args)
			must.Nil(err)

			rets := make([]interface{}, len(ret))
			for i := range ret {
				rets[i] = ret[i].Interface()
			}
			should.Equal(test.ret, rets)
		})
	}
}

func TestNamespaceAPI(t *testing.T) {
	// handlers := make(map[string]*namespaceHandler)
	// broadcast := NewBroadcast()
	// conn := &conn{
	// 	//TODO:
	// 	Conn: engineio.Conn,

	// 	broadcast:  broadcast,
	// 	encoder:    parser.NewEncoder(c),
	// 	decoder:    parser.NewDecoder(c),
	// 	errorChan:  make(chan errorMessage),
	// 	writeChan:  make(chan writePacket),
	// 	quitChan:   make(chan struct{}),
	// 	handlers:   handlers,
	// 	namespaces: make(map[string]*namespaceConn),
	// }
	// nc := newNamespaceConn(conn, header.Namespace, conn.braodcast)
	// context :=
	// nc.SetContext()

	// savedContext := nc.Context()
	// if savedContext != context {
	// 	t.Errorf("[]")
	// }

}

func TestErrorNamespaceAPI(t *testing.T) {

}

func TestNamespaceDispacth(t *testing.T) {

}
