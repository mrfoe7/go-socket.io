package parser

import (
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"testing"
)

func TestDecoder(t *testing.T) {
	at := assert.New(t)

	var decoder Reader
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

