package asm

import (
	"bytes"
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

func TestParserSized(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{42}, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 8 {
		t.Fatalf("expected 0 byte write count, got %v", n)
	}
	rb := r.Bytes()
	if !bytes.Equal(rb, []byte{0x00, vm.LOAD, 0x03, 0x66, 0x6f, 0x6f, 0x01, 0x2a}) {
		t.Fatalf("expected 0x00%x012a, got %v", vm.LOAD, rb)
	}
}
