package asm

import (
	"bytes"
	"encoding/hex"
	"log"
	"testing"

	"git.defalsify.org/vise.git/vm"
)

func TestParserRoute(t *testing.T) {
	b := bytes.NewBuffer(nil)
	s := "HALT\n"
	Parse(s, b)
	expect := vm.NewLine(nil, vm.HALT, nil, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)	
	}

	b = bytes.NewBuffer(nil)
	s = "MSINK\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MSINK, nil, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "MAP tinkywinky\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MAP, []string{"tinkywinky"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)	
	}

	b = bytes.NewBuffer(nil)
	s = "MOVE dipsy\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MOVE, []string{"dipsy"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
			log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "RELOAD lalapu\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.RELOAD, []string{"lalapu"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "LOAD foo 42\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.LOAD, []string{"foo"}, []byte{0x2a}, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "MOUT foo bar\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MOUT, []string{"foo", "bar"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "MOUT baz 42\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MOUT, []string{"baz", "42"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "MNEXT inky 12\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MNEXT, []string{"inky", "12"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "MPREV pinky 34\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MPREV, []string{"pinky", "34"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "INCMP foo bar\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.INCMP, []string{"foo", "bar"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "INCMP baz 42\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.INCMP, []string{"baz", "42"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "INCMP xyzzy *\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.INCMP, []string{"xyzzy", "*"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "DOWN foo 2 bar\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MOUT, []string{"bar", "2"}, nil, nil)
	expect = vm.NewLine(expect, vm.HALT, nil, nil, nil)
	expect = vm.NewLine(expect, vm.INCMP, []string{"foo", "2"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "UP 3 bar\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MOUT, []string{"bar", "3"}, nil, nil)
	expect = vm.NewLine(expect, vm.HALT, nil, nil, nil)
	expect = vm.NewLine(expect, vm.INCMP, []string{"_", "3"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "NEXT 4 baz\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MNEXT, []string{"baz", "4"}, nil, nil)
	expect = vm.NewLine(expect, vm.HALT, nil, nil, nil)
	expect = vm.NewLine(expect, vm.INCMP, []string{">", "4"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

	b = bytes.NewBuffer(nil)
	s = "PREVIOUS 5 xyzzy\n"
	Parse(s, b)
	expect = vm.NewLine(nil, vm.MPREV, []string{"xyzzy", "5"}, nil, nil)
	expect = vm.NewLine(expect, vm.HALT, nil, nil, nil)
	expect = vm.NewLine(expect, vm.INCMP, []string{"<", "5"}, nil, nil)
	if !bytes.Equal(b.Bytes(), expect) {
		log.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, b)
	}

}

func TestParserInit(t *testing.T) {
	var b []byte
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b = vm.NewLine(b, vm.CATCH, []string{"xyzzy"}, []byte{0x02, 0x9a}, []uint8{1})
	b = vm.NewLine(b, vm.INCMP, []string{"pinky", "inky"}, nil, nil)
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
	b = vm.NewLine(b, vm.MOUT, []string{"foo", "baz_ba_zbaz"}, nil, nil)
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
	expect := []byte{0x00, vm.MOUT, 0x03, 0x66, 0x6f, 0x6f, 0x0b, 0x62, 0x61, 0x7a, 0x5f, 0x62, 0x61, 0x5f, 0x7a, 0x62, 0x61, 0x7a}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected:\n\t%x\ngot:\n\t%x", expect, rb)
	}
}

func TestParseDouble(t *testing.T) {
	t.Skip("foo")
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
	expect := []byte{0x00, vm.INCMP, 0x03, 0x62, 0x61, 0x72, 0x03, 0x66, 0x6f, 0x6f}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected %x, got %x", expect, rb)
	}
}

func TestParseMenu(t *testing.T) {
	s := `DOWN foobar 00 inky_pinky
UP s1 tinkywinky
UP 2 dipsy
`
	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("wrote %v bytes", n)

	s = `MOUT inky_pinky 00
MOUT tinkywinky s1
MOUT dipsy 2
HALT
INCMP foobar 00
INCMP _ s1
INCMP _ 2
`
	r_check := bytes.NewBuffer(nil)
	n, err = Parse(s, r_check)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("wrote %v bytes", n)

	if !bytes.Equal(r_check.Bytes(), r.Bytes()) {
		t.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", r_check, r)
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
	b := vm.NewLine(nil, vm.CATCH, []string{"plugh"}, []byte{0x02, 0x9a}, []uint8{0x2a})
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

	b = vm.NewLine(nil, vm.CATCH, []string{"plugh"}, []byte{0x01}, []uint8{0x0})
	s, err = vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r = bytes.NewBuffer(nil)
	n, err = Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 11 {
		t.Fatalf("expected 11 byte write count, got %v", n)
	}
	rb = r.Bytes()
	expect_hex = "000105706c756768010100"
	expect, err = hex.DecodeString(expect_hex)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rb, expect) {
		t.Fatalf("expected %v, got %x", expect_hex, rb)
	}
}

func TestParseCroak(t *testing.T) {
	b := bytes.NewBuffer(nil)
	s := "CROAK 2 1\n"
	Parse(s, b)
	expect := vm.NewLine(nil, vm.CROAK, nil, []byte{0x02}, []uint8{0x1})
	if !bytes.Equal(b.Bytes(), expect) {
		t.Fatalf("expected %x, got %x", expect, b)
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
	b = vm.NewLine(b, vm.INCMP, []string{"pinky", "inky"}, nil, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{42}, nil)
	b = vm.NewLine(b, vm.MOUT, []string{"bar", "bar_barb_az"}, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s\n", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("result %x", r.Bytes())

	r_expect_hex := "000700010578797a7a7902029a0100080570696e6b7904696e6b79000303666f6f012a000a036261720b6261725f626172625f617a"
	r_expect, err := hex.DecodeString(r_expect_hex)
	if err != nil {
		t.Fatal(err)
	}
	n_expect := len(r_expect)
	if n != n_expect {
		t.Fatalf("expected total %v bytes output, got %v", n_expect, n)
	}
	
	rb := r.Bytes()
	if !bytes.Equal(rb, r_expect) {
		t.Fatalf("expected result:\n\t%v, got:\n\t%x", r_expect_hex, rb)
	}

	_, err = vm.ParseAll(rb, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParserCapQuote(t *testing.T) {
	t.Skip("please fix mysterious ignore of initial cap in display sym match")
	b := vm.NewLine(nil, vm.MOUT, []string{"a", "foo"}, nil, nil) 
	b = vm.NewLine(b, vm.MOUT, []string{"b", "Bar"}, nil, nil) 
	b = vm.NewLine(b, vm.MOUT, []string{"c", "baz"}, nil, nil) 
	b = vm.NewLine(b, vm.MSINK, nil, nil, nil)
	s, err := vm.ToString(b)
	log.Printf("parsing:\n%s", s)

	r := bytes.NewBuffer(nil)
	n, err := Parse(s, r)
	if err != nil {
		t.Fatal(err)
	}
	_ = n
}
