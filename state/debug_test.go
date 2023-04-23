package state

import (
	"fmt"
	"testing"
)

func TestDebugFlagDenied(t *testing.T) {
	err := FlagDebugger.Register(7, "foo")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestDebugFlagString(t *testing.T) {
	err := FlagDebugger.Register(8, "FOO")
	if err != nil {
		t.Fatal(err)
	}
	err = FlagDebugger.Register(9, "BAR")
	if err != nil {
		t.Fatal(err)
	}
	err = FlagDebugger.Register(11, "BAZ")
	if err != nil {
		t.Fatal(err)
	}
	flags := []byte{0x06, 0x09}
	r := FlagDebugger.AsString(flags, 4)
	expect := "INTERNAL_INMATCH(1),INTERNAL_TERMINATE(2),FOO(8),BAZ(11)" 
	if r != expect {
		t.Fatalf("expected '%s', got '%s'", expect, r)
	}
}

func TestDebugState(t *testing.T) {
	err := FlagDebugger.Register(8, "FOO")
	if err != nil {
		t.Fatal(err)
	}
	st := NewState(1).WithDebug()
	st.SetFlag(FLAG_DIRTY)
	st.SetFlag(8)
	st.Down("root")

	r := fmt.Sprintf("%s", st)
	expect := "moves: 1 idx: 0 flags: INTERNAL_DIRTY(3),FOO(8) path: root lang: (default)"
	if r != expect {
		t.Fatalf("expected '%s', got '%s'", expect, r)
	}
}
