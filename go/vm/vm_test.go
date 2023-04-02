package vm

import (
	"testing"
)

func TestParseNoArg(t *testing.T) {
	b := NewLine(nil, HALT, nil, nil, nil)
	b, err := ParseHalt(b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseSym(t *testing.T) {
	b := NewLine(nil, MAP, []string{"baz"}, nil, nil)
	sym, b, err := ParseMap(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "baz" {
		t.Fatalf("expected sym baz, got %v", sym)
	}

	b = NewLine(nil, RELOAD, []string{"xyzzy"}, nil, nil)
	sym, b, err = ParseReload(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "xyzzy" {
		t.Fatalf("expected sym xyzzy, got %v", sym)
	}

	b = NewLine(nil, MOVE, []string{"plugh"}, nil, nil)
	sym, b, err = ParseMove(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "plugh" {
		t.Fatalf("expected sym plugh, got %v", sym)
	}
}

func TestParseTwoSym(t *testing.T) {
	b := NewLine(nil, INCMP, []string{"foo", "bar"}, nil, nil)
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
}

func TestParseSymSig(t *testing.T) {
	b := NewLine(nil, CATCH, []string{"baz"}, nil, []uint8{0x0d})
	sym, n, b, err := ParseCatch(b)
	if err != nil {
		t.Fatal(err)
	}
	if sym != "baz" {
		t.Fatalf("expected sym baz, got %v", sym)
	}
	if n != 13 {
		t.Fatalf("expected n 13, got %v", n)
	}
}

func TestParseSymAndLen(t *testing.T) {
	b := NewLine(nil, LOAD, []string{"foo"}, []byte{0x2a}, nil)
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

	b = NewLine(nil, LOAD, []string{"baz"}, []byte{0x0}, nil)
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
}
