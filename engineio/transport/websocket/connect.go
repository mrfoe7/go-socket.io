package websocket

import (
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/googollee/go-socket.io/engineio/transport"
)

// conn implements base.Conn
type conn struct {
	transport.FrameReader
	transport.FrameWriter

	wsWrapper wrapper

	url          url.URL
	remoteHeader http.Header

	closed    chan struct{}
	closeOnce sync.Once
}

func newConn(ws *websocket.Conn, url url.URL, header http.Header) *conn {
	w := newWrapper(ws)

	return &conn{
		url:          url,
		remoteHeader: header,
		wsWrapper:    w,
		closed:       make(chan struct{}),
		FrameReader:  packet.NewDecoder(w),
		FrameWriter:  packet.NewEncoder(w),
	}
}

func (c *conn) URL() url.URL {
	return c.url
}

func (c *conn) RemoteHeader() http.Header {
	return c.remoteHeader
}

func (c *conn) LocalAddr() net.Addr {
	return c.wsWrapper.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.wsWrapper.RemoteAddr()
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.wsWrapper.SetReadDeadline(t)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	// TODO: is locking really needed for SetWriteDeadline? If so, what about
	// the read deadline?
	c.wsWrapper.writeLocker.Lock()
	err := c.wsWrapper.SetWriteDeadline(t)
	c.wsWrapper.writeLocker.Unlock()

	return err
}

func (c *conn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	<-c.closed
}

func (c *conn) Close() error {
	c.closeOnce.Do(func() {
		close(c.closed)
	})
	return c.wsWrapper.Close()
}
