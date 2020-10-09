package parser

import (
	"io"

	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/googollee/go-socket.io/engineio/utils"
)

type decoder struct {
	r FrameReader
}

func newDecoder(r FrameReader) *decoder {
	return &decoder{
		r: r,
	}
}

func (e *decoder) NextReader() (frame.Type, packet.Type, io.ReadCloser, error) {
	ft, r, err := e.r.FrameRead()
	if err != nil {
		return frame.String, packet.OPEN, nil, err
	}

	var b[1]byte

	if _, err := io.ReadFull(r, b[:]); err != nil {
		_ = r.Close()

		return frame.String, packet.OPEN, nil, err
	}

	return ft, utils.ByteToPacketType(b[0], ft), r, nil
}
