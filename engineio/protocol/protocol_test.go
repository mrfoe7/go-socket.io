package protocol

import (
	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/googollee/go-socket.io/engineio/utils"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Frame struct {
	typ  frame.Type
	data []byte
}

type Packet struct {
	ft   frame.Type
	pt   packet.Type
	data []byte
}

var tests = []struct {
	packets []Packet
	frames  []Frame
}{
	{nil, nil},
	{[]Packet{
			{frame.String, packet.OPEN, []byte{}},
		},
		[]Frame{
			{frame.String, []byte("0")}},
	},
	{[]Packet{
		{frame.String, packet.MESSAGE, []byte("hello 你好")},
	}, []Frame{
		{frame.String, []byte("4hello 你好")}},
	},
	{[]Packet{
		{frame.Binary, packet.MESSAGE, []byte("hello 你好")},
	}, []Frame{
		{frame.Binary, []byte{0x04, 'h', 'e', 'l', 'l', 'o', ' ', 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}},
		},
	},
	{[]Packet{
		{frame.String, packet.OPEN, []byte{}},
		{frame.Binary, packet.MESSAGE, []byte("hello\n")},
		{frame.String, packet.MESSAGE, []byte("你好\n")},
		{frame.String, packet.PING, []byte("probe")},
	},
	[]Frame{
			{frame.String, []byte("0")},
			{frame.Binary, []byte{0x04, 'h', 'e', 'l', 'l', 'o', '\n'}},
			{frame.String, []byte("4你好\n")},
			{frame.String, []byte("2probe")},
		},
	},
	{[]Packet{
		{frame.Binary, packet.MESSAGE, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
		{frame.String, packet.MESSAGE, []byte("hello")},
		{frame.String, packet.CLOSE, []byte{}},
	}, []Frame{
		{frame.Binary, []byte{4, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
		{frame.String, []byte("4hello")},
		{frame.String, []byte("1")},
	}},
}

func TestEncoder(t *testing.T) {
	at := assert.New(t)

	var writer *fakeConnWriter
	var encoder utils.Writer

	for _, test := range tests {
		writer = newFakeConnWriter()

		encoder = NewEncoder(writer)

		for _, p := range test.packets {
			fw, err := encoder.NextWriter(p.ft, p.pt)
			at.Nil(err)

			_, err = fw.Write(p.data)
			at.Nil(err)

			err = fw.Close()
			at.Nil(err)
		}

		at.Equal(test.frames, writer.frames)
	}
}

func TestDecoder(t *testing.T) {
	at := assert.New(t)

	var decoder utils.Reader
	for _, test := range tests {
		decoder = NewDecoder(newFakeConnReader(test.frames))

		var output []Packet

		for {
			ft, pt, fr, err := decoder.NextReader()
			if err != nil {
				at.Equal(io.EOF, err)
				break
			}

			b, err := ioutil.ReadAll(fr)
			at.Nil(err)

			fr.Close()

			output = append(output, Packet{
				ft:   ft,
				pt:   pt,
				data: b,
			})
		}

		at.Equal(test.packets, output)
	}
}

func BenchmarkEncoder(b *testing.B) {
	encoder := NewEncoder(&fakeDiscardWriter{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w, _ := encoder.NextWriter(frame.String, packet.MESSAGE)
		w.Close()
		w, _ = encoder.NextWriter(frame.Binary, packet.MESSAGE)
		w.Close()
	}
}

func BenchmarkDecoder(b *testing.B) {
	r := newFakeConstReader()
	decoder := NewDecoder(r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, fr, _ := decoder.NextReader()
		fr.Close()

		_, _, fr, _ = decoder.NextReader()
		fr.Close()
	}
}
