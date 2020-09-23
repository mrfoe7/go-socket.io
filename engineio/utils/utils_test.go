package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
)

func TestPacketType(t *testing.T) {
	at := assert.New(t)

	tests := []struct {
		b         byte
		frameType frame.Type
		typ       packet.Type
		strbyte   byte
		binbyte   byte
		str       string
	}{
		{0, frame.Binary, packet.OPEN, '0', 0, "open"},
		{1, frame.Binary, packet.CLOSE, '1', 1, "close"},
		{2, frame.Binary, packet.PING, '2', 2, "ping"},
		{3, frame.Binary, packet.PONG, '3', 3, "pong"},
		{4, frame.Binary, packet.MESSAGE, '4', 4, "message"},
		{5, frame.Binary, packet.UPGRADE, '5', 5, "upgrade"},
		{6, frame.Binary, packet.NOOP, '6', 6, "noop"},

		{'0', frame.String, packet.OPEN, '0', 0, "open"},
		{'1', frame.String, packet.CLOSE, '1', 1, "close"},
		{'2', frame.String, packet.PING, '2', 2, "ping"},
		{'3', frame.String, packet.PONG, '3', 3, "pong"},
		{'4', frame.String, packet.MESSAGE, '4', 4, "message"},
		{'5', frame.String, packet.UPGRADE, '5', 5, "upgrade"},
		{'6', frame.String, packet.NOOP, '6', 6, "noop"},
	}

	for _, test := range tests {
		typ := ByteToPacketType(test.b, test.frameType)

		at.Equal(test.typ, typ)
		at.Equal(test.strbyte, typ.StringByte())
		at.Equal(test.binbyte, typ.BinaryByte())
		at.Equal(test.str, typ.String())
		at.Equal(test.str, (typ).String())
	}
}