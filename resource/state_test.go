package resource

import (
	"testing"

	"git.defalsify.org/vise/state"
)

func TestStateResourceInit(t *testing.T) {
	st := state.NewState(0)
	rs := NewMenuResource()
	_ = ToStateResource(rs).WithState(&st)
	_ = NewStateResource(&st)
}
