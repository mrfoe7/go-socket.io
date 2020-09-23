package engineio

import (
	"github.com/googollee/go-socket.io/engineio/transport"
	"net/http"
	"time"

	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

// ReqCheckerFunc
type ReqCheckerFunc func(*http.Request) (http.Header, error)

// ConnInitFunc
type ConnInitFunc func(*http.Request, transport.Conn)

// Options is options to create a server.
type Options struct {
	//
	SessionIDGenerator SessionIDGenerator

	//
	Transports         []transport.Transporter

	//
	PingTimeout        time.Duration

	//
	PingInterval       time.Duration

	//
	ReqChecker     ReqCheckerFunc

	//
	ConnInitor      ConnInitFunc
}

func defaultChecker(*http.Request) (http.Header, error) {
	return nil, nil
}

func defaultInitor(*http.Request, transport.Conn) {}

func newServerOption() *Options{
	return &Options{
		SessionIDGenerator: newDefaultIDGenerator(),
		Transports: []transport.Transporter{
			polling.Default,
			websocket.Default,
		},
		PingInterval: time.Second * 20,
		PingTimeout: time.Minute,
		ReqChecker: defaultChecker,
		ConnInitor: defaultInitor,
	}
}

func (o *Options) getRequestChecker() ReqCheckerFunc {
	return o.ReqChecker
}

func (o *Options) getConnectInitor() ConnInitFunc {
	return o.ConnInitor
}

func (c *Options) getPingTimeout() time.Duration {
	return c.PingTimeout
}

func (c *Options) getPingInterval() time.Duration {
	return c.PingInterval
}

func (o *Options) getTransport() []transport.Transporter {
	return o.Transports
}

func (o *Options) getSessionIDGenerator() SessionIDGenerator {
	return o.SessionIDGenerator
}

func (o *Options) getServerConfig() *serverConfig{
	return &serverConfig{
		pingInterval: o.getPingInterval(),
		pingTimeout: o.getPingTimeout(),

		requestChecker: o.getRequestChecker(),
		connectInitor: o.getConnectInitor(),
	}
}

type serverConfig struct {
	pingInterval   time.Duration
	pingTimeout    time.Duration

	requestChecker ReqCheckerFunc
	connectInitor  ConnInitFunc
}
