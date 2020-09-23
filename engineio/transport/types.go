package transport

import (
	"errors"
	"net/http"
	"net/url"
)

// Transporter is a transport which can creates base.Conn
type Transporter interface {
	Name() string
	Accept(http.ResponseWriter, *http.Request) (Conn, error)
	Dial(url.URL, http.Header) (Conn, error)
}

// Checker is function to check request.
type Checker func(*http.Request) (http.Header, error)

var (
	// ErrInvalidFrame is returned when writing invalid frame type.
	ErrInvalidFrame = errors.New("invalid frame type")

	// ErrInvalidPacket is returned when writing invalid packet type.
	ErrInvalidPacket = errors.New("invalid packet type")
)
