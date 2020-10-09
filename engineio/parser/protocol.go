// Package packet is codec of packet for connection which supports framing.
package parser

import (
	"github.com/googollee/go-socket.io/engineio/packet"
	"io"

	"github.com/googollee/go-socket.io/engineio/frame"
)

type Parser interface {
	FrameReader
	FrameWriter
}

// Reader read frame to packet. It need be closed before next reading.
type Reader interface {
	NextReader() (frame.Type, packet.Type, io.ReadCloser, error)
}

// Writer write frame to packet. It need be closed before next writing.
type Writer interface {
	NextWriter(frame.Type, packet.Type) (io.WriteCloser, error)
}

// FrameReader is the reader which supports framing.
type FrameReader interface {
	FrameRead() (frame.Type, io.ReadCloser, error)
}

// FrameWriter is the writer which supports framing.
type FrameWriter interface {
	FrameWrite(frame.Type) (io.WriteCloser, error)
}

// NewEncoder creates a packet encoder which writes to w.
func NewEncoder(w FrameWriter) Writer {
	return newEncoder(w)
}

// NewDecoder creates a packet decoder which reads from r.
func NewDecoder(r FrameReader) Reader {
	return newDecoder(r)
}
