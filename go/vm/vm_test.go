package vm

import (
	"context"
	"fmt"
	"testing"
	
	"git.defalsify.org/festive/state"
)

type TestResource struct {
}

func (r *TestResource) Get(sym string) (string, error) {
	switch sym {
	case "foo":
		return "inky pinky blinky clyde", nil
	case "bar":
		return "inky pinky {.one} blinky {.two} clyde", nil
	}
	return "", fmt.Errorf("unknown symbol %s", sym)
}

func (r *TestResource) Render(sym string, values map[string]string) (string, error) {
	v, err := r.Get(sym)
	return v, err
}

func TestRun(t *testing.T) {
	st := state.NewState(5, 255)
	rs := TestResource{}
	b := []byte{0x00, 0x02}
	r, err := Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Errorf("error on valid opcode: %v", err)	
	}

	b = []byte{0x01, 0x02}
	r, err = Run(b, st, &rs, context.TODO())
	if err == nil {
		t.Errorf("no error on invalid opcode")	
	}
	_ = r
}
