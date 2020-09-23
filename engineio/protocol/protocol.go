// Package packet is codec of packet for connection which supports framing.
package protocol

import (
	"io"

	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/utils"
)

// FrameReader is the reader which supports framing.
type FrameReader interface {
	NextReader() (frame.Type, io.ReadCloser, error)
}

// FrameWriter is the writer which supports framing.
type FrameWriter interface {
	NextWriter(frame.Type) (io.WriteCloser, error)
}

// NewEncoder creates a packet encoder which writes to w.
func NewEncoder(w FrameWriter) utils.Writer {
	return newEncoder(w)
}

// NewDecoder creates a packet decoder which reads from r.
func NewDecoder(r FrameReader) utils.Reader {
	return newDecoder(r)
}
