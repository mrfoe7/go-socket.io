package frame

// FrameType is the type of frames.
type Type byte

const (
	// FrameString identifies a string frame.
	String Type = iota

	// FrameBinary identifies a binary frame.
	Binary
)

// ByteToFrameType converts a byte to FrameType.
func ByteToFrameType(b byte) Type {
	return Type(b)
}

// Byte returns type in byte.
func (t Type) Byte() byte {
	return byte(t)
}
