package asm

import (
	"log"
	"testing"

	"git.defalsify.org/festive/vm"
)


func TestParserInit(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.CATCH, []string{"xyzzy"}, []byte{0x02, 0x9a}, []uint8{1})
	b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{42}, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"bar", "barbarbaz"}, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)
	n, err := Parse(s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected 0 byte write count, got %v", n)
	}
}
