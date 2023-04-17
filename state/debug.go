package state

import (
	"fmt"
	"strings"
)

type flagDebugger struct {
	flagStrings map[uint32]string
}

func newFlagDebugger() flagDebugger {
	fd := flagDebugger{
		flagStrings: make(map[uint32]string),
	}
	fd.register(FLAG_READIN, "INTERNAL_READIN")
	fd.register(FLAG_INMATCH, "INTERNAL_INMATCH")
	fd.register(FLAG_TERMINATE, "INTERNAL_TERMINATE")
	fd.register(FLAG_DIRTY, "INTERNAL_DIRTY")
	fd.register(FLAG_WAIT, "INTERNAL_WAIT")
	fd.register(FLAG_LOADFAIL, "INTERNAL_LOADFAIL")
	return fd
}

func(fd *flagDebugger) register(flag uint32, name string) {
	fd.flagStrings[flag] = name
}

func(fd *flagDebugger) Register(flag uint32, name string) error {
	if (flag < 8) {
		return fmt.Errorf("flag %v is not definable by user", flag)
	}
	fd.register(flag, name)
	return nil
}

func(fd *flagDebugger) AsString(flags []byte, length uint32) string {
	var r []string
	var i uint32
	for i = 0; i < length + 8; i++ {
		if getFlag(i, flags) {
			s := fmt.Sprintf("%s(%v)", fd.flagStrings[i], i)
			r = append(r, s)
		}
	}
	return strings.Join(r, ",")
}

var (
	FlagDebugger = newFlagDebugger()
)
