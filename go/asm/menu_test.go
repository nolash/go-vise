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
	err = m.Add("NEXT", "1", "pinky", "bar")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("PREVIOUS", "2", "blinky clyde", "baz")
	if err != nil {
		t.Fatal(err)
	}
	err = m.Add("UP", "99", "tinky-winky", "xyzzy")
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
MOUT 1 "pinky"
MOUT 2 "blinky clyde"
MOUT 99 "tinky-winky"
HALT
INCMP 0 foo
INCMP 99 _
`
	if r != expect {
		t.Errorf("expected:\n\t%v\ngot:\n\t%v\n", expect, r)
	}
}
