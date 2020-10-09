package websocket

import (
	"github.com/googollee/go-socket.io/engineio/parser"
	"github.com/googollee/go-socket.io/engineio/transport"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type conn struct {
	parser.Writer
	parser.Reader

	url          url.URL
	remoteHeader http.Header
	ws           wrapper

	closed       chan struct{}

	closeOnce    sync.Once
}

func newConn(wsConn *websocket.Conn, url *url.URL, header http.Header) transport.Conn {
	w := newWrapper(wsConn)

	return &conn{
		url:          *url,
		remoteHeader: header,
		ws:           w,
		closed:       make(chan struct{}),
		//todo: mb usage parser without protocol
		Reader: parser.NewDecoder(w),
		Writer: parser.NewEncoder(w),
	}
}

func (c *conn) URL() url.URL {
	return c.url
}

func (c *conn) RemoteHeader() http.Header {
	return c.remoteHeader
}

func (c *conn) LocalAddr() net.Addr {
	return c.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.RemoteAddr()
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.SetReadDeadline(t)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	// TODO: is locking really needed for SetWriteDeadline? If so, what about
	// the read deadline?
	c.ws.writeLocker.Lock()
	err := c.ws.SetWriteDeadline(t)
	c.ws.writeLocker.Unlock()

	return err
}

func (c *conn) ServeHTTP(http.ResponseWriter, *http.Request) {
	<-c.closed
}

func (c *conn) Close() error {
	c.closeOnce.Do(func() {
		close(c.closed)
	})

	return c.ws.Close()
}

func (c *conn) Read(b []byte) (n int, err error) {
	panic("implement me")
}

func (c *conn) Write(b []byte) (n int, err error) {
	panic("implement me")
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.SetDeadline(t)
}
