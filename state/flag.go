package state

const (
	FLAG_READIN = iota
	FLAG_INMATCH 
	FLAG_DIRTY
	FLAG_WAIT
	FLAG_LOADFAIL
	FLAG_TERMINATE 
	FLAG_BLOCK
	FLAG_LANG
	FLAG_USERSTART = 8
)

func IsWriteableFlag(flag uint32) bool {
	if flag > 5 {
		return true
	}
	//if flag & FLAG_WRITEABLE > 0 {
	//	return true	
	//}
	return false
}

// Retrieve the state of a state flag
func getFlag(bitIndex uint32, bitField []byte) bool {
	byteIndex := bitIndex / 8
	localBitIndex := bitIndex % 8
	b := bitField[byteIndex]
	return (b & (1 << localBitIndex)) > 0
}
