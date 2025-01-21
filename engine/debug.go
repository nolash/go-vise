package engine

import (
	"fmt"
	"io"
	"os"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
)

// Debug implementations output details about the execution state on each execution halt.
type Debug interface {
	// Break receives the state and cache in order to generate its output.
	Break(*state.State, cache.Memory)
}

// SimpleDebug is a vanilla implementation of the Debug interface.
type SimpleDebug struct {
	pfx string
	w   io.Writer
}

// NewSimpleDebug instantiates a new SimpleDebug object.
func NewSimpleDebug(w io.Writer) Debug {
	if w == nil {
		w = os.Stderr
	}
	return &SimpleDebug{
		w:   w,
		pfx: "DUMP>",
	}
}

// Break implements the Debug interface.
func (dbg *SimpleDebug) Break(st *state.State, ca cache.Memory) {
	fmt.Fprintf(dbg.w, "%s State:\n", dbg.pfx)
	node, lvl := st.Where()
	fmt.Fprintf(dbg.w, "%s\tPath: %s (%d)\n", dbg.pfx, node, lvl)
	fmt.Fprintf(dbg.w, "%s\tFlags:\n", dbg.pfx)
	for _, s := range state.FlagDebugger.AsList(st.Flags, st.BitSize-8) {
		fmt.Fprintf(dbg.w, "%s\t\t%s\n", dbg.pfx, s)
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
