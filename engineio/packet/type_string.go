// Code generated by "stringer -type=Type"; DO NOT EDIT.

package packet

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[OPEN-0]
	_ = x[CLOSE-1]
	_ = x[PING-2]
	_ = x[PONG-3]
	_ = x[MESSAGE-4]
	_ = x[UPGRADE-5]
	_ = x[NOOP-6]
}

const _Type_name = "OPENCLOSEPINGPONGMESSAGEUPGRADENOOP"

var _Type_index = [...]uint8{0, 4, 9, 13, 17, 24, 31, 35}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
