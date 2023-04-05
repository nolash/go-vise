package asm

import (
	"bytes"
	"encoding/hex"
	"log"
	"testing"

	"git.defalsify.org/festive/vm"
)


func TestParserInit(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.CATCH, []string{"xyzzy"}, []byte{0x02, 0x9a}, []uint8{1})
	b = vm.NewLine(b, vm.INCMP, []string{"inky", "pinky"}, nil, nil)
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
		t.Fatalf("expected 8 byte write count, got %v", n)
	}
	rb := r.Bytes()
	if !bytes.Equal(rb, []byte{0x00, vm.LOAD, 0x03, 0x66, 0x6f, 0x6f, 0x01, 0x2a}) {
		t.Fatalf("expected 0x00%x012a, got %v", vm.LOAD, rb)
	}
}

func TestParseDisplay(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.MOUT, []string{"foo", "baz ba zbaz"}, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 18 {
		t.Fatalf("expected 18 byte write count, got %v", n)
	}
	rb := r.Bytes()
	expect := []byte{0x00, vm.MOUT, 0x03, 0x66, 0x6f, 0x6f, 0x0b, 0x62, 0x61, 0x7a, 0x20, 0x62, 0x61, 0x20, 0x7a, 0x62, 0x61, 0x7a}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected %x, got %x", expect, rb)
	}
}

func TestParseDouble(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.INCMP, []string{"foo", "bar"}, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 10 {
		t.Fatalf("expected 18 byte write count, got %v", n)
	}
	rb := r.Bytes()
	expect := []byte{0x00, vm.INCMP, 0x03, 0x66, 0x6f, 0x6f, 0x03, 0x62, 0x61, 0x72}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected %x, got %x", expect, rb)
	}
}

func TestParseSingle(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.MAP, []string{"xyzzy"}, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 8 {
		t.Fatalf("expected 8 byte write count, got %v", n)
	}
	rb := r.Bytes()
	expect := []byte{0x00, vm.MAP, 0x05, 0x78, 0x79, 0x7a, 0x7a, 0x79}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected %x, got %x", expect, rb)
	}
}

func TestParseSig(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.CATCH, []string{"plugh"}, []byte{0x02, 0x9a}, []uint8{0x2a})
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 12 {
		t.Fatalf("expected 12 byte write count, got %v", n)
	}
	rb := r.Bytes()
	expect_hex := "000105706c75676802029a01"
	expect, err := hex.DecodeString(expect_hex)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected %v, got %x", expect_hex, rb)
	}
}

func TestParseNoarg(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 8 byte write count, got %v", n)
	}
	rb := r.Bytes()
	expect := []byte{0x00, vm.HALT}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected %x, got %x", expect, rb)
	}
}

func TestParserWriteMultiple(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.CATCH, []string{"xyzzy"}, []byte{0x02, 0x9a}, []uint8{1})
	b = vm.NewLine(b, vm.INCMP, []string{"inky", "pinky"}, nil, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{42}, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"bar", "bar barb az"}, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	n_expect := 2 // halt
	n_expect += 2 + 6 + 3 + 1 // catch
	n_expect += 2 + 5 + 6 // incmp
	n_expect += 2 + 4 + 2 // load
	n_expect += 2 + 4 + 12 // mout
	log.Printf("result %x", r.Bytes())
	if n != n_expect {
		t.Fatalf("expected total %v bytes output, got %v", n_expect, n)
	}
	r_expect_hex := "000700010578797a7a7902029a000804696e6b790570696e6b79000303666f6f023432000a036261720b626172206261726220617a"
	r_expect, err := hex.DecodeString(r_expect_hex)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(r.Bytes(), r_expect) {
		t.Fatalf("expected result %v, got %x", r_expect_hex, r.Bytes())
	}
}
