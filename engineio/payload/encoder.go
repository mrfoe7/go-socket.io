package payload

import (
	"bytes"
	"encoding/base64"
	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"io"
)

type writerFeeder interface {
	getWriter() (io.Writer, error)
	putWriter(error) error
}

type Encoder struct {
	supportBinary bool
	feeder        writerFeeder

	ft         frame.Type
	pt         packet.Type
	header     bytes.Buffer
	frameCache bytes.Buffer
	b64Writer  io.WriteCloser
	rawWriter  io.Writer
}

func (e *Encoder) NOOP() []byte {
	if e.supportBinary {
		return []byte{0x00, 0x01, 0xff, '6'}
	}
	return []byte("1:6")
}

func (e *Encoder) NextWriter(ft frame.Type, pt packet.Type) (io.WriteCloser, error) {
	w, err := e.feeder.getWriter()
	if err != nil {
		return nil, err
	}
	e.rawWriter = w

	e.ft = ft
	e.pt = pt
	e.frameCache.Reset()

	if !e.supportBinary && ft == frame.Binary {
		e.b64Writer = base64.NewEncoder(base64.StdEncoding, &e.frameCache)
	} else {
		e.b64Writer = nil
	}
	return e, nil
}

func (e *Encoder) Write(p []byte) (int, error) {
	if e.b64Writer != nil {
		return e.b64Writer.Write(p)
	}
	return e.frameCache.Write(p)
}

type writeHeaderFunc func() error

func (e *Encoder) Close() error {
	if e.b64Writer != nil {
		e.b64Writer.Close()
	}

	var writeHeader writeHeaderFunc
	writeHeader = e.writeBinaryHeader

	if !e.supportBinary {
		writeHeader = e.writeTextHeader

		if e.ft == frame.Binary {
			writeHeader = e.writeB64Header
		}
	}

	e.header.Reset()

	err := writeHeader()
	if err == nil {
		_, err = e.header.WriteTo(e.rawWriter)
	}
	if err == nil {
		_, err = e.frameCache.WriteTo(e.rawWriter)
	}
	if writeErr := e.feeder.putWriter(err); writeErr != nil {
		return writeErr
	}

	return err
}

func (e *Encoder) writeTextHeader() error {
	l := int64(e.frameCache.Len() + 1) // length for packet type
	err := writeTextLen(l, &e.header)
	if err == nil {
		err = e.header.WriteByte(e.pt.StringByte())
	}
	return err
}

func (e *Encoder) writeB64Header() error {
	l := int64(e.frameCache.Len() + 2) // length for 'b' and packet type
	err := writeTextLen(l, &e.header)
	if err == nil {
		err = e.header.WriteByte('b')
	}
	if err == nil {
		err = e.header.WriteByte(e.pt.StringByte())
	}

	return err
}

func (e *Encoder) writeBinaryHeader() error {
	l := int64(e.frameCache.Len() + 1) // length for packet type
	b := e.pt.StringByte()

	if e.ft == frame.Binary {
		b = e.pt.BinaryByte()
	}
	err := e.header.WriteByte(e.ft.Byte())
	if err == nil {
		err = writeBinaryLen(l, &e.header)
	}
	if err == nil {
		err = e.header.WriteByte(b)
	}

	return err
}
