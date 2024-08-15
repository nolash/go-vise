package engine

import (
	"fmt"
	"io"
	"os"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
)

type Debug interface {
	Break(*state.State, cache.Memory)
}

type SimpleDebug struct {
	pfx string
	w io.Writer
}

func NewSimpleDebug(w io.Writer) Debug {
	if w == nil {
		w = os.Stderr	
	} 
	return &SimpleDebug{
		w: w,
		pfx: "DUMP>",
	}
}

func (dbg* SimpleDebug) Break(st *state.State, ca cache.Memory) {
	fmt.Fprintf(dbg.w, "%s State:\n", dbg.pfx)
	for _, s := range state.FlagDebugger.AsList(st.Flags, st.BitSize - 8) {
		fmt.Fprintf(dbg.w, "%s\t%s\n", dbg.pfx, s)
	}
	for i := uint32(0); i < ca.Levels(); i++ {
		fmt.Fprintf(dbg.w, "%s Cache[%d]:\n", dbg.pfx, i)
		ks := ca.Keys(i)
		for _, k := range ks {
			v, err := ca.Get(k)
			if err != nil {
				continue
			}
			fmt.Fprintf(dbg.w, "%s\t%s: %v\n", dbg.pfx, k, v)
		}
	}
}
