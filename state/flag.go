package state

const (
	FLAG_READIN = iota
	FLAG_INMATCH 
	FLAG_TERMINATE 
	FLAG_DIRTY
	FLAG_WAIT
	FLAG_LOADFAIL
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

type FlagDebugger struct {
}
