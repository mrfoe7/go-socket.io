package payload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpError(t *testing.T) {
	tests := []struct {
		op        OpType
		err       error
		temporary bool
		errMsg string
	}{
		{read, errPaused, true, "read: paused"},
		{read, errTimeout, false, "read: timeout"},
	}

	var err error

	for _, test := range tests {
		err = newOpError(test.op, test.err)

		assert.EqualError(t, err, test.errMsg)

		re, ok := err.(PayloadError)
		assert.True(t, ok)

		assert.Equal(t, test.temporary, re.Temporary())
	}
}
