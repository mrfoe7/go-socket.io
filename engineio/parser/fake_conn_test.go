package parser

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
)

type fakeConnReader struct {
	frames []Frame
}

func newFakeConnReader(frames []Frame) *fakeConnReader {
	return &fakeConnReader{
		frames: frames,
	}
}

func (r *fakeConnReader) FrameRead() (frame.Type, io.ReadCloser, error) {
	if len(r.frames) == 0 {
		return frame.String, nil, io.EOF
	}
	f := r.frames[0]
	r.frames = r.frames[1:]

	return f.typ, ioutil.NopCloser(bytes.NewReader(f.data)), nil
}

type fakeFrame struct {
	w    *fakeConnWriter
	typ  frame.Type
	data *bytes.Buffer
}

func newFakeFrame(w *fakeConnWriter, typ frame.Type) *fakeFrame {
	return &fakeFrame{
		w:    w,
		typ:  typ,
		data: bytes.NewBuffer(nil),
	}
}

func (w *fakeFrame) Write(p []byte) (int, error) {
	return w.data.Write(p)
}

func (w *fakeFrame) Read(p []byte) (int, error) {
	return w.data.Read(p)
}

func (w *fakeFrame) Close() error {
	if w.w == nil {
		return nil
	}
	w.w.frames = append(w.w.frames, Frame{
		typ:  w.typ,
		data: w.data.Bytes(),
	})
	return nil
}

type fakeConnWriter struct {
	frames []Frame
}

func newFakeConnWriter() *fakeConnWriter {
	return &fakeConnWriter{}
}

func (w *fakeConnWriter)FrameWrite(typ frame.Type) (io.WriteCloser, error) {
	return newFakeFrame(w, typ), nil
}

type fakeOneFrameConst struct {
	b byte
}

func (c *fakeOneFrameConst) Read(p []byte) (int, error) {
	p[0] = c.b
	return 1, nil
}

type fakeConstReader struct {
	ft frame.Type
	r  *fakeOneFrameConst
}

func newFakeConstReader() *fakeConstReader {
	return &fakeConstReader{
		ft: frame.String,
		r: &fakeOneFrameConst{
			b: packet.MESSAGE.StringByte(),
		},
	}
}

func (r *fakeConstReader) FrameRead() (frame.Type, io.ReadCloser, error) {
	switch r.ft {
	case frame.Binary:
		r.ft = frame.String
		r.r.b = packet.MESSAGE.StringByte()
	case frame.String:
		r.ft = frame.Binary
		r.r.b = packet.MESSAGE.BinaryByte()
	}

	return r.ft, ioutil.NopCloser(r.r), nil
}

type fakeOneFrameDiscarder struct{}

func (d fakeOneFrameDiscarder) Write(p []byte) (int, error) {
	return len(p), nil
}

func (d fakeOneFrameDiscarder) Close() error {
	return nil
}

type fakeDiscardWriter struct{}

func (w *fakeDiscardWriter) FrameWrite(frame.Type) (io.WriteCloser, error) {
	return fakeOneFrameDiscarder{}, nil
}
