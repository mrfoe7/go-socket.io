package transport

import (
	"encoding/json"
	"io"
	"net"
	"net/url"
	"time"

	"github.com/googollee/go-socket.io/engineio/utils"
)

// Conn is interface a connection.
type Conn interface {
	net.Conn

	utils.Reader
	utils.Writer

	URL() url.URL
}

// ConnParams is connection parameter of server.
type ConnParams struct {
	PingInterval time.Duration
	PingTimeout  time.Duration
	SID          string
	Upgrades     []string
}

type jsonParameters struct {
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
}

// ReadConnParameters reads ConnParameters from r.
func ReadConnParameters(r io.Reader) (ConnParams, error) {
	var param jsonParameters

	if err := json.NewDecoder(r).Decode(&param); err != nil {
		return ConnParams{}, err
	}

	return ConnParams{
		SID:          param.SID,
		Upgrades:     param.Upgrades,
		PingInterval: time.Duration(param.PingInterval) * time.Millisecond,
		PingTimeout:  time.Duration(param.PingTimeout) * time.Millisecond,
	}, nil
}

// WriteTo writes to w with json format.
func (p ConnParams) WriteTo(w io.Writer) (int64, error) {
	arg := jsonParameters{
		SID:          p.SID,
		Upgrades:     p.Upgrades,
		PingInterval: int(p.PingInterval / time.Millisecond),
		PingTimeout:  int(p.PingTimeout / time.Millisecond),
	}

	writer := connParamsWriter{
		w: w,
	}

	err := json.NewEncoder(&writer).Encode(arg)

	return writer.i, err
}

type connParamsWriter struct {
	i int64
	w io.Writer
}

func (w *connParamsWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.i += int64(n)

	return n, err
}
