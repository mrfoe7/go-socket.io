package parser

import (
	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncoder(t *testing.T) {
	at := assert.New(t)

	var writer *fakeConnWriter
	var encoder Writer

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

func BenchmarkEncoder(b *testing.B) {
	encoder := NewEncoder(&fakeDiscardWriter{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w, _ := encoder.NextWriter(frame.String, packet.MESSAGE)
		_ = w.Close()

		w, _ = encoder.NextWriter(frame.Binary, packet.MESSAGE)
		_ = w.Close()
	}
}