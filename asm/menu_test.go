package asm

import (
	"testing"

	"git.defalsify.org/vise.git/vm"
)

func TestMenuInterpreter(t *testing.T) {
	m := NewMenuProcessor()
	ph := vm.NewParseHandler().WithDefaultHandlers()
	err := m.Add("DOWN", "0", "inky", "foo")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("NEXT", "1", "pinky", "")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("PREVIOUS", "2", "blinkyclyde", "")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("UP", "99", "tinky_winky", "")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("BOGUS", "42", "lala poo", "plugh")
	if err == nil {
		t.Errorf("expected error on invalid menu item 'BOGUS'")
	}
	b := m.ToLines()
	r, err := ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect := `MOUT inky 0
MNEXT pinky 1
MPREV blinkyclyde 2
MOUT tinky_winky 99
HALT
INCMP foo 0
INCMP > 1
INCMP < 2
INCMP _ 99
`
	if r != expect {
		t.Errorf("expected:\n\t%v\ngot:\n\t%v\n", expect, r)
	}
}
