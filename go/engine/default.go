package engine

import (
	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/resource"
)

func NewDefaultEngine(dir string) Engine {
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	return NewEngine(&st, &rs)
}
