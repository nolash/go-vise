package vm

import (
	"testing"
)

func TestParseNoArg(t *testing.T) {
	b := NewLine(nil, HALT, nil, nil, nil)
	_, b, _ = opSplit(b)
	b, err := ParseHalt(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}
}

func TestParseSym(t *testing.T) {
	b := NewLine(nil, MAP, []string{"baz"}, nil, nil)
	_, b, _ = opSplit(b)
	sym, b, err := ParseMap(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "baz" {
		t.Fatalf("expected sym baz, got %v", sym)
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}

	b = NewLine(nil, RELOAD, []string{"xyzzy"}, nil, nil)
	_, b, _ = opSplit(b)
	sym, b, err = ParseReload(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "xyzzy" {
		t.Fatalf("expected sym xyzzy, got %v", sym)
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}

	b = NewLine(nil, MOVE, []string{"plugh"}, nil, nil)
	_, b, _ = opSplit(b)
	sym, b, err = ParseMove(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "plugh" {
		t.Fatalf("expected sym plugh, got %v", sym)
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}
}

func TestParseTwoSym(t *testing.T) {
	b := NewLine(nil, INCMP, []string{"foo", "bar"}, nil, nil)
	_, b, _ = opSplit(b)
	one, two, b, err := ParseInCmp(b)
	if err != nil {
		t.Fatal(err)
	}
	if one != "foo" {
		t.Fatalf("expected symone foo, got %v", one)
	}
	if two != "bar" {
		t.Fatalf("expected symtwo bar, got %v", two)
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}
}

func TestParseSig(t *testing.T) {
	b := NewLine(nil, CROAK, nil, []byte{0x0b, 0x13}, []uint8{0x04})
	_, b, _ = opSplit(b)
	n, m, b, err := ParseCroak(b)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2835 {
		t.Fatalf("expected n 13, got %v", n)
	}
	if !m {
		t.Fatalf("expected m true")
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}
}

func TestParseSymSig(t *testing.T) {
	b := NewLine(nil, CATCH, []string{"baz"}, []byte{0x0a, 0x13}, []uint8{0x01})
	_, b, _ = opSplit(b)
	sym, n, m, b, err := ParseCatch(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "baz" {
		t.Fatalf("expected sym baz, got %v", sym)
	}
	if n != 2579 {
		t.Fatalf("expected n 13, got %v", n)
	}
	if !m {
		t.Fatalf("expected m true")
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}
}

func TestParseSymAndLen(t *testing.T) {
	b := NewLine(nil, LOAD, []string{"foo"}, []byte{0x2a}, nil)
	_, b, _ = opSplit(b)
	sym, n, b, err := ParseLoad(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "foo" {
		t.Fatalf("expected sym foo, got %v", sym)
	}
	if n != 42 {
		t.Fatalf("expected n 42, got %v", n)
	}

	b = NewLine(nil, LOAD, []string{"bar"}, []byte{0x02, 0x9a}, nil)
	_, b, _ = opSplit(b)
	sym, n, b, err = ParseLoad(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "bar" {
		t.Fatalf("expected sym foo, got %v", sym)
	}
	if n != 666 {
		t.Fatalf("expected n 666, got %v", n)
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}

	b = NewLine(nil, LOAD, []string{"baz"}, []byte{0x0}, nil)
	_, b, _ = opSplit(b)
	sym, n, b, err = ParseLoad(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "baz" {
		t.Fatalf("expected sym foo, got %v", sym)
	}
	if n != 0 {
		t.Fatalf("expected n 666, got %v", n)
	}
	if len(b) > 0 {
		t.Fatalf("expected empty code")
	}
}
