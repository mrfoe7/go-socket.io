package utils

import (
	"io"
	"time"

	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
)

var chars = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_")

// Timestamp returns a string based on different nano time.
func Timestamp() string {
	now := time.Now().UnixNano()
	ret := make([]byte, 0, 16)

	for now > 0 {
		ret = append(ret, chars[int(now%int64(len(chars)))])
		now = now / int64(len(chars))
	}

	return string(ret)
}

// ByteToPacketType converts a byte to packet.Type.
func ByteToPacketType(b byte, ft frame.Type) packet.Type {
	if ft == frame.String {
		b -= '0'
	}

	return packet.Type(b)
}

// Reader reads a frame. It need be closed before next reading.
type Reader interface {
	NextReader() (frame.Type, packet.Type, io.ReadCloser, error)
}

// Writer writes a frame. It need be closed before next writing.
type Writer interface {
	NextWriter(frame.Type, packet.Type) (io.WriteCloser, error)
}
