package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/utils"
)

func TestPacketType(t *testing.T) {
	at := assert.New(t)

	tests := []struct {
		b         byte
		frameType frame.Type
		typ       Type
		strbyte   byte
		binbyte   byte
		str       string
	}{
		{0, frame.Binary, OPEN, '0', 0, "open"},
		{1, frame.Binary, CLOSE, '1', 1, "close"},
		{2, frame.Binary, PING, '2', 2, "ping"},
		{3, frame.Binary, PONG, '3', 3, "pong"},
		{4, frame.Binary, MESSAGE, '4', 4, "message"},
		{5, frame.Binary, UPGRADE, '5', 5, "upgrade"},
		{6, frame.Binary, NOOP, '6', 6, "noop"},

		{'0', frame.String, OPEN, '0', 0, "open"},
		{'1', frame.String, CLOSE, '1', 1, "close"},
		{'2', frame.String, PING, '2', 2, "ping"},
		{'3', frame.String, PONG, '3', 3, "pong"},
		{'4', frame.String, MESSAGE, '4', 4, "message"},
		{'5', frame.String, UPGRADE, '5', 5, "upgrade"},
		{'6', frame.String, NOOP, '6', 6, "noop"},
	}

	for _, test := range tests {
		typ := utils.ByteToPacketType(test.b, test.frameType)

		at.Equal(test.typ, typ)
		at.Equal(test.strbyte, typ.StringByte())
		at.Equal(test.binbyte, typ.BinaryByte())
		at.Equal(test.str, typ.String())
		at.Equal(test.str, (typ).String())
	}
}
