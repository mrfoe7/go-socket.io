package socketio

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

type fakeConn struct {
	Id     int
	output io.Writer
	lock   sync.RWMutex
}

func (fc *fakeConn) ID() string {
	fc.lock.RLock()
	fc.Id++
	fc.lock.RUnlock()

	return strconv.Itoa(fc.Id)
}

func (fc *fakeConn) Emit(message string, v ...interface{}) {
	output := new(bytes.Buffer)
	if l := len(v); l > 0 {
		last := v[l-1]
		lastV := reflect.ValueOf(last)
		if lastV.Kind() == reflect.Func {
			v = v[:l-1]
		}
		for _, value := range v {
			valueV := reflect.ValueOf(value)
			output.Write(valueV.Bytes())
		}
	}
	_, err := fc.output.Write(output.Bytes())
	if err != nil {
		return
	}
}

type FakeConnPool struct {
}

func TestBroadcastNewBroadcast(t *testing.T) {
	expectedBroadcast := &broadcast{rooms: make(map[string]map[string]BroadcastConn)}
	broadcast := NewBroadcast()
	if !reflect.DeepEqual(broadcast, expectedBroadcast) {
		t.Errorf("Wrong result, expected %#v, got %#v", expectedBroadcast, broadcast)
	}
}

func TestBroadcastAPI(t *testing.T) {
	out := new(bytes.Buffer)
	BuffersPool := []*bytes.Buffer{
		new(bytes.Buffer),
		new(bytes.Buffer),
		new(bytes.Buffer),
	}
	broadcast := NewBroadcast()
	RoomsPool := []string{
		"room1",
		"room2",
		"room3",
	}
	FakeConnPool := []fakeConn{
		fakeConn{},
		fakeConn{},
		fakeConn{},
	}
	for _, room := range RoomsPool {
		for _, conn := range FakeConnPool {
			broadcast.Join(room, conn)
		}

		if broadcast.Len(room) != 3 {

		}
		broadcast.Clear(room)
		if broadcast.Len(room) != 0 {

		}
		result := out.String()
	}
	// func TestBroadcastJoin(t *testing.T) {

	// }

	// func TestBroadcastLeave(t *testing.T) {

	// }

	// func TestBroadcastLeaveAll(t *testing.T) {

	// }

	// func TestBroadcastClear(t *testing.T) {

	// }

	// func TestBroadcastSend(t *testing.T) {

	// }

	// func TestBroadcastSendAll(t *testing.T) {

	// }

	// func TestBroadcastLen(t *testing.T) {

	// }

	// func TestBroadcastRooms(t *testing.T) {

	// }

}
