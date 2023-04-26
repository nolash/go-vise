package asm

import (
	"testing"

	"git.defalsify.org/vise.git/vm"
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
	r, err := vm.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect := `MOUT 0 inky
MNEXT 1 pinky
MPREV 2 blinkyclyde
MOUT 99 tinky_winky
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
