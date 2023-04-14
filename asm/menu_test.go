package asm

import (
	"testing"

	"git.defalsify.org/festive/vm"
)


func TestMenuInterpreter(t *testing.T) {
	m := NewMenuProcessor()
	err := m.Add("DOWN", "0", "inky", "foo")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("NEXT", "1", "pinky", "")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("PREVIOUS", "2", "blinky clyde", "")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("UP", "99", "tinky-winky", "")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("BOGUS", "42", "lala poo", "plugh")
	if err == nil {
		t.Errorf("expected error on invalid menu item 'BOGUS'")
	}
	b := m.ToLines()
	r, err := vm.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect := `MOUT 0 "inky"
MNEXT 1 "pinky"
MPREV 2 "blinky clyde"
MOUT 99 "tinky-winky"
HALT
INCMP 0 foo
INCMP 1 >
INCMP 2 <
INCMP 99 _
`
	if r != expect {
		t.Errorf("expected:\n\t%v\ngot:\n\t%v\n", expect, r)
	}
}