package payload

import (
	"bufio"
	"encoding/base64"
	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/googollee/go-socket.io/engineio/utils"
	"io"
	"io/ioutil"
)

type byteReader interface {
	io.Reader
	ReadByte() (byte, error)
}

type readerFeeder interface {
	getReader() (io.Reader, bool, error)
	putReader(error) error
}

type decoder struct {
	limitReader io.LimitedReader
	b64Reader   io.Reader
	feeder      readerFeeder
	rawReader   byteReader

	supportBinary bool

	ft frame.Type
	pt packet.Type
}

func (d *decoder) NextReader() (frame.Type, packet.Type, io.ReadCloser, error) {
	if d.rawReader == nil {
		r, supportBinary, err := d.feeder.getReader()
		if err != nil {
			return 0, 0, nil, err
		}
		br, ok := r.(byteReader)
		if !ok {
			br = bufio.NewReader(r)
		}
		if err := d.setNextReader(br, supportBinary); err != nil {
			return 0, 0, nil, d.sendError(err)
		}
	}

	return d.ft, d.pt, d, nil
}

func (d *decoder) Read(p []byte) (int, error) {
	if d.b64Reader != nil {
		return d.b64Reader.Read(p)
	}
	return d.limitReader.Read(p)
}

func (d *decoder) Close() error {
	if _, err := io.Copy(ioutil.Discard, d); err != nil {
		return d.sendError(err)
	}
	err := d.setNextReader(d.rawReader, d.supportBinary)

	if err != nil {
		if err != io.EOF {
			return d.sendError(err)
		}

		d.rawReader = nil
		d.limitReader.R = nil
		d.limitReader.N = 0
		d.b64Reader = nil

		err = d.sendError(nil)
	}
	return err
}

func (d *decoder) setNextReader(r byteReader, supportBinary bool) error {
	var read func(byteReader) (frame.Type, packet.Type, int64, error)

	if supportBinary {
		read = d.binaryRead
	} else {
		read = d.textRead
	}

	ft, pt, l, err := read(r)
	if err != nil {
		return err
	}

	d.ft = ft
	d.pt = pt
	d.rawReader = r
	d.limitReader.R = r
	d.limitReader.N = l
	d.supportBinary = supportBinary

	d.b64Reader = nil

	if !supportBinary && ft == frame.Binary {
		d.b64Reader = base64.NewDecoder(base64.StdEncoding, &d.limitReader)
	}

	return nil
}

func (d *decoder) sendError(err error) error {
	if e := d.feeder.putReader(err); e != nil {
		return e
	}

	return err
}

func (d *decoder) textRead(r byteReader) (frame.Type, packet.Type, int64, error) {
	l, err := readTextLen(r)
	if err != nil {
		return 0, 0, 0, err
	}

	ft := frame.String
	b, err := r.ReadByte()
	if err != nil {
		return 0, 0, 0, err
	}
	l--

	if b == 'b' {
		ft = frame.Binary
		b, err = r.ReadByte()
		if err != nil {
			return 0, 0, 0, err
		}
		l--
	}

	pt := utils.ByteToPacketType(b, frame.String)
	return ft, pt, l, nil
}

func (d *decoder) binaryRead(r byteReader) (frame.Type, packet.Type, int64, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, 0, 0, err
	}
	if b > 1 {
		return 0, 0, 0, errInvalidPayload
	}
	ft := frame.ByteToFrameType(b)

	l, err := readBinaryLen(r)
	if err != nil {
		return 0, 0, 0, err
	}

	b, err = r.ReadByte()
	if err != nil {
		return 0, 0, 0, err
	}
	pt := utils.ByteToPacketType(b, ft)
	l--

	return ft, pt, l, nil
}
