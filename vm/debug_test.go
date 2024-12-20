package vm

import (
	"testing"
)


func TestToString(t *testing.T) {
	var b []byte
	var r string
	var expect string
	var err error

	ph := NewParseHandler().WithDefaultHandlers()
	b = NewLine(nil, CATCH, []string{"xyzzy"}, []byte{0x0d}, []uint8{1})
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "CATCH xyzzy 13 1\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, CROAK, nil, []byte{0x0d}, []uint8{1})
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "CROAK 13 1\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, LOAD, []string{"foo"}, []byte{0x0a}, nil)
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "LOAD foo 10\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, RELOAD, []string{"bar"}, nil, nil)
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "RELOAD bar\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, MAP, []string{"inky_pinky"}, nil, nil) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "MAP inky_pinky\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, MOVE, []string{"blinky_clyde"}, nil, nil) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "MOVE blinky_clyde\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, HALT, nil, nil, nil) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "HALT\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, INCMP, []string{"13", "baz"}, nil, nil) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect = "INCMP 13 baz\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, MNEXT, []string{"11", "nextmenu"}, nil, nil) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	//expect = "MNEXT 11 \"nextmenu\"\n"
	expect = "MNEXT 11 nextmenu\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, MPREV, []string{"222", "previous menu item"}, nil, nil) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	//expect = "MPREV 222 \"previous menu item\"\n"
	expect = "MPREV 222 previous menu item\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, MOUT, []string{"1", "foo"}, nil, nil) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	//expect = "MOUT 1 \"foo\"\n"
	expect = "MOUT 1 foo\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}

	b = NewLine(nil, MSINK, nil, nil, nil) //[]uint8{0x42, 0x2a}) 
	r, err = ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	//expect = "MSIZE 66 42\n"
	expect = "MSINK\n"
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}
}

func TestToStringMultiple(t *testing.T) {
	ph := NewParseHandler().WithDefaultHandlers()
	b := NewLine(nil, INCMP, []string{"1", "foo"}, nil, nil)
	b = NewLine(b, INCMP, []string{"2", "bar"}, nil, nil)
	b = NewLine(b, CATCH, []string{"aiee"}, []byte{0x02, 0x9a}, []uint8{0})
	b = NewLine(b, LOAD, []string{"inky"}, []byte{0x2a}, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	r, err := ph.ToString(b)
	if err != nil {
		t.Fatal(err)
	}
	expect := `INCMP 1 foo
INCMP 2 bar
CATCH aiee 666 0
LOAD inky 42
HALT
`
	if r != expect {
		t.Fatalf("expected:\n\t%v\ngot:\n\t%v", expect, r)
	}
}

func TestVerifyMultiple(t *testing.T) {
	ph := NewParseHandler().WithDefaultHandlers()
	b := NewLine(nil, INCMP, []string{"1", "foo"}, nil, nil)
	b = NewLine(b, INCMP, []string{"2", "bar"}, nil, nil)
	b = NewLine(b, CATCH, []string{"aiee"}, []byte{0x02, 0x9a}, []uint8{0})
	b = NewLine(b, LOAD, []string{"inky"}, []byte{0x2a}, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	n, err := ph.ParseAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected write count to be 0, was %v (how is that possible)", n)
	}
}
