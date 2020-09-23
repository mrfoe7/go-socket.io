package polling

import (
	"github.com/googollee/go-socket.io/engineio/todo"
	"github.com/googollee/go-socket.io/engineio/transport"
	"net/http"
	"net/url"
	"time"
)

// CheckOriginFunc
type CheckOriginFunc func(r *http.Request) bool

// Transport is the transport of polling.
type Transport struct {
	Client      *http.Client
	CheckOrigin CheckOriginFunc
}

// Default is the default transport.
var Default = &Transport{
	Client: &http.Client{
		Timeout: time.Minute,
	},
	CheckOrigin: nil,
}

// Name is the name of transport.
func (t *Transport) Name() string {
	return "polling"
}

// Accept accepts a http request and create Conn.
func (t *Transport) Accept(_ http.ResponseWriter, r *http.Request) (transport.Conn, error) {
	return newConn(t, r), nil
}

// Dial dials connection to url.
func (t *Transport) Dial(u url.URL, requestHeader http.Header) (transport.Conn, error) {
	query := u.Query()
	query.Set("transport", t.Name())
	u.RawQuery = query.Encode()

	client := t.Client
	if client == nil {
		client = Default.Client
	}

	return todo.dial(client, u, requestHeader)
}
