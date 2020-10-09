package payload

import (
	"errors"
	"fmt"
)

// PayloadError is payload error.
// todo: ref error name
type PayloadError interface {
	Error() string
	Temporary() bool
}

// OpError is operation error.
type OpError struct {
	Op  OpType
	Err error
}

func newOpError(op OpType, err error) error {
	return &OpError{
		Op:  op,
		Err: err,
	}
}

func (e *OpError) Error() string {
	return fmt.Sprintf("%s: %s", e.Op, e.Err.Error())
}

// Temporary returns true if error can retry.
func (e *OpError) Temporary() bool {
	if pe, ok := e.Err.(PayloadError); ok {
		return pe.Temporary()
	}
	return false
}

type retryError struct {
	err string
}

func (e retryError) Error() string {
	return e.err
}

func (e retryError) Temporary() bool {
	return true
}

var (
	errPaused = retryError{"paused"}

	errTimeout = errors.New("timeout")

	errInvalidPayload = errors.New("invalid payload")

	errDrain = errors.New("drain")

	errOverlap = errors.New("overlap")
)
