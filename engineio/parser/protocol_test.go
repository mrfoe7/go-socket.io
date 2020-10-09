package parser

import (
	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
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
