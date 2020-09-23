package websocket

import (
	"fmt"
	"github.com/googollee/go-socket.io/engineio/frame"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/gorilla/websocket"
)

type wrapper struct {
	*websocket.Conn

	writeLocker *sync.Mutex
	readLocker  *sync.Mutex
}

func newWrapper(conn *websocket.Conn) wrapper {
	return wrapper{
		Conn:        conn,
		writeLocker: new(sync.Mutex),
		readLocker:  new(sync.Mutex),
	}
}

func (w wrapper) NextReader() (frame.Type, io.ReadCloser, error) {
	w.readLocker.Lock()
	typ, r, err := w.Conn.NextReader()
	// The wrapper remains locked until the returned ReadCloser is Closed.
	if err != nil {
		w.readLocker.Unlock()
		return 0, nil, err
	}

	switch typ {
	case websocket.TextMessage:
		return frame.String, newRcWrapper(w.readLocker, r), nil
	case websocket.BinaryMessage:
		return frame.Binary, newRcWrapper(w.readLocker, r), nil
	}

	w.readLocker.Unlock()
	return 0, nil, transport.ErrInvalidFrame
}

type rcWrapper struct {
	io.Reader
	nagTimer *time.Timer

	quitNag  chan struct{}

	lock        *sync.Mutex
}

func newRcWrapper(lock *sync.Mutex, r io.Reader) rcWrapper {
	timer := time.NewTimer(30 * time.Second)
	q := make(chan struct{})

	go func() {
		select {
		case <-q:
		case <-timer.C:
			fmt.Println("Did you forget to Close() the ReadCloser from NextReader?")
		}
	}()

	return rcWrapper{
		nagTimer: timer,
		quitNag:  q,
		lock:        lock,
		Reader:   r,
	}
}

func (r rcWrapper) Close() error {
	// Stop the nagger.
	r.nagTimer.Stop()

	close(r.quitNag)

	// Attempt to drain the Reader.
	// reader may be closed, ignore error
	io.Copy(ioutil.Discard, r)

	// Unlock the wrapper's read lock for future calls to NextReader.
	r.lock.Unlock()

	return nil
}

func (w wrapper) NextWriter(typ frame.Type) (io.WriteCloser, error) {
	var t int

	switch typ {
	case frame.String:
		t = websocket.TextMessage
	case frame.Binary:
		t = websocket.BinaryMessage
	default:
		return nil, transport.ErrInvalidFrame
	}

	w.writeLocker.Lock()
	writer, err := w.Conn.NextWriter(t)

	// The wrapper remains locked until the returned WriteCloser is Closed.
	if err != nil {
		w.writeLocker.Unlock()
		return nil, err
	}

	return newWcWrapper(w.writeLocker, writer), nil
}

type wcWrapper struct {
	io.WriteCloser

	nagTimer *time.Timer

	quitNag  chan struct{}

	lock        *sync.Mutex
}

func newWcWrapper(lock *sync.Mutex, w io.WriteCloser) wcWrapper {
	timer := time.NewTimer(30 * time.Second)
	q := make(chan struct{})

	go func() {
		select {
		case <-q:
		case <-timer.C:
			fmt.Println("Did you forget to Close() the WriteCloser from NextWriter?")
		}
	}()

	return wcWrapper{
		nagTimer:    timer,
		quitNag:     q,
		lock:           lock,
		WriteCloser: w,
	}
}

func (w wcWrapper) Close() error {
	// Unlock the wrapper's write lock for future calls to NextWriter.
	defer w.lock.Unlock()

	// Stop the nagger.
	w.nagTimer.Stop()

	close(w.quitNag)

	return w.WriteCloser.Close()
}
