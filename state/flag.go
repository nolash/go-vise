package state

const (
	// Currently reading input. Set by first INCMP instruction encontered.
	FLAG_READIN = iota
	// Input matched a selector. Set by first INCMP matching input.
	FLAG_INMATCH
	// The instruction HALT has been encountered.
	FLAG_WAIT
	// The last LOAD or RELOAD executed returneded an error.
	FLAG_LOADFAIL
	// A LOAD or RELOAD has returned fresh data.
	FLAG_DIRTY
	// Not currently in use.
	FLAG_RESERVED
	// VM execution is blocked.
	FLAG_TERMINATE
	// The return value from a LOAD or RELOAD is a new language selection.
	FLAG_LANG
	// User-defined flags start here.
	FLAG_USERSTART = 8
)

const (
	nonwriteable_flag_threshold = FLAG_RESERVED
)

// IsWriteableFlag returns true if flag can be set by implementer code.
func IsWriteableFlag(flag uint32) bool {
	if flag > nonwriteable_flag_threshold {
		return true
	}
	return false
}

// Retrieve the state of a state flag
func getFlag(bitIndex uint32, bitField []byte) bool {
	byteIndex := bitIndex / 8
	localBitIndex := bitIndex % 8
	b := bitField[byteIndex]
	return (b & (1 << localBitIndex)) > 0
}
