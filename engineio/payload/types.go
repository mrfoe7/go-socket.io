//go:generate stringer -type=OpType

package payload

type OpType int

const (
	read OpType = iota

	write

	payload
)
