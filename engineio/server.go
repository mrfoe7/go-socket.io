package engineio

import (
	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/googollee/go-socket.io/engineio/transport"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Server is server.
type Server struct {
	transports TransportManager
	sessions   SessionManager

	config *serverConfig

	connChan  chan transport.Conn

	closeOnce sync.Once
}

// NewServer returns a server.
func NewServer(opts *Options) *Server {
	if opts == nil {
		opts = newServerOption()
	}

	return &Server{
		transports: newTransportManager(opts.getTransport()),
		sessions:   newSessionManager(opts.getSessionIDGenerator()),
		config:     opts.getServerConfig(),

		connChan: make(chan transport.Conn, 1),
	}
}

// Close closes server.
func (s *Server) Close() error {
	s.closeOnce.Do(func() {
		close(s.connChan)
	})

	return nil
}

// Accept accepts a connection.
func (s *Server) Accept() (transport.Conn, error) {
	c := <-s.connChan

	if c == nil {
		return nil, io.EOF
	}

	return c, nil
}

//
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	sid := query.Get("sid")
	session := s.sessions.Get(sid)

	transportType := query.Get("transport")
	trspt := s.transports.Get(transportType)

	if trspt == nil {
		http.Error(w, "unsupported transport type", http.StatusBadRequest)
		return
	}

	header, err := s.config.requestChecker(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	for k, v := range header {
		w.Header()[k] = v
	}

	if session == nil {
		if sid != "" {
			http.Error(w, "invalid sid", http.StatusBadRequest)
			return
		}
		conn, err := trspt.Accept(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		params := transport.ConnParams{
			PingInterval: s.config.pingInterval,
			PingTimeout:  s.config.pingTimeout,
			Upgrades:     s.transports.UpgradeFrom(transportType),
			SID:          s.sessions.NewID(),
		}

		session, err = newSession(conn, s.sessions, params, transportType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		s.config.connectInitor(r, session)

		go func() {
			w, err := session.nextWriter(frame.String, packet.OPEN)
			if err != nil {
				session.Close()
				return
			}

			if _, err := session.params.WriteTo(w); err != nil {
				w.Close()
				session.Close()

				return
			}

			if err := w.Close(); err != nil {
				session.Close()

				return
			}

			s.connChan <- session
		}()
	}

	if session.Transport() != transportType {
		conn, err := trspt.Accept(w, r)
		if err != nil {
			// don't call http.Error() for HandshakeErrors because
			// they get handled by the websocket library internally.
			if _, ok := err.(websocket.HandshakeError); !ok {
				http.Error(w, err.Error(), http.StatusBadGateway)
			}
			return
		}

		session.upgrade(transportType, conn)
		if handler, ok := conn.(http.Handler); ok {
			handler.ServeHTTP(w, r)
		}

		return
	}

	session.serveHTTP(w, r)
}

// Count counts connected
func (s *Server) Count() int {
	return s.sessions.Count()
}
