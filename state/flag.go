package state

const (
	FLAG_READIN = 1
	FLAG_INMATCH = 2
	FLAG_TERMINATE = 3
	FLAG_DIRTY = 4
	FLAG_LOADFAIL = 5
)

func IsWriteableFlag(flag uint32) bool {
	if flag > 7 {
		return true
	}
	//if flag & FLAG_WRITEABLE > 0 {
	//	return true	
	//}
	return false
}
