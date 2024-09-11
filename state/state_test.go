package state

import (
	"bytes"
	"testing"
)

// Check creation 
func TestNewState(t *testing.T) {
	st := NewState(5)
	if len(st.Flags) != 2 {
		t.Fatalf("invalid state flag length: %v", len(st.Flags))
	}
	st = NewState(8)
	if len(st.Flags) != 2 {
		t.Fatalf("invalid state flag length: %v", len(st.Flags))
	}
	st = NewState(17)
	if len(st.Flags) != 4 {
		t.Fatalf("invalid state flag length: %v", len(st.Flags))
	}
	v := st.FlagBitSize()
	x := uint32(17+8)
	if v != x {
		t.Fatalf("expected %d, get %d", x, v)
	}
	v = uint32(st.FlagByteSize())
	x = 4
	if v != x {
		t.Fatalf("expected %d, get %d", x, v)
	}
	if !IsWriteableFlag(8) {
		t.Fatal("expected true")
	}
}

func TestStateflags(t *testing.T) {
	st := NewState(9)
	v := st.GetFlag(2)
	if v {
		t.Fatalf("Expected bit 2 not to be set")
	}
	v = st.SetFlag(2)
	if !v {
		t.Fatalf("Expected change to be set for bit 2")
	}
	v = st.GetFlag(2)
	if !v {
		t.Fatalf("Expected bit 2 to be set")
	}
	v = st.SetFlag(10)
	if !v {
		t.Fatalf("Expected change to be set for bit 10")
	}
	v = st.GetFlag(10)
	if !v {
		t.Fatalf("Expected bit 10 to be set")
	}
	v = st.ResetFlag(2)
	if !v {
		t.Fatalf("Expected change to be set for bit 10")
	}
	v = st.GetFlag(2)
	if v {
		t.Fatalf("Expected bit 2 not to be set")
	}
	v = st.GetFlag(10)
	if !v {
		t.Fatalf("Expected bit 10 to be set")
	}
	v = st.SetFlag(10)
	if v {
		t.Fatalf("Expected change not to be set for bit 10")
	}
	v = st.SetFlag(2)
	v = st.SetFlag(16)
	if !bytes.Equal(st.Flags[:3], []byte{0x04, 0x04, 0x01}) {
		t.Fatalf("Expected 0x040401, got %v", st.Flags[:3])
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	v = st.SetFlag(42)
}

func TestStateFlagReset(t *testing.T) {
	st := NewState(9)
	v := st.SetFlag(10)
	v = st.SetFlag(11)
	v = st.ResetFlag(10)
	if !v {
		t.Fatal("expected true")
	}
	v = st.ResetFlag(10)
	if v {
		t.Fatal("expected false")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	v = st.ResetFlag(42)
}

func TestStateFlagGetOutOfRange(t *testing.T) {
	st := NewState(1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	st.GetFlag(9)
}

func TestStateFlagFromSlice(t *testing.T) {
	st := NewState(15)
	st.SetFlag(2)
	v := st.GetIndex([]byte{})
	if v {
		t.Fatalf("Expected no match on empty compare")
	}
	v = st.GetIndex([]byte{0x01})
	if v {
		t.Fatalf("Expected 0x01 not to match")
	}
	v = st.GetIndex([]byte{0x04})
	if !v {
		t.Fatalf("Expected 0x04 to match")
	}
	st.SetFlag(12)
	v = st.GetIndex([]byte{0x04})
	if !v {
		t.Fatalf("Expected 0x04 to match")
	}
	v = st.GetIndex([]byte{0x00, 0x10})
	if !v {
		t.Fatalf("Expected 0x1000 to match")
	}
	v = st.ResetFlag(2)
	v = st.GetIndex([]byte{0x00, 0x10})
	if !v {
		t.Fatalf("Expected 0x1000 to matck")
	}
}

func TestStateNavigate(t *testing.T) {
	st := NewState(0)
	err := st.Down("foo")
	if err != nil {
		t.Fatal(err)
	}
	err = st.Down("bar")
	if err != nil {
		t.Fatal(err)
	}
	err = st.Down("baz")
	if err != nil {
		t.Fatal(err)
	}

	s, i := st.Where()
	if s != "baz" {
		t.Fatalf("expected baz, got %s", s)
	}
	if i != 0 {
		t.Fatalf("expected idx 0, got %v", i)
	}
	r := st.Depth()
	if r != 2 {
		t.Fatalf("expected depth 3, got %v", r)
	}

	s, err = st.Up()
	if err != nil {
		t.Fatal(err)
	}
	if s != "bar" {
		t.Fatalf("expected bar, got %s", s)
	}
	s, i = st.Where()
	if s != "bar" {
		t.Fatalf("expected bar, got %s", s)
	}
	if i != 0 {
		t.Fatalf("expected idx 0, got %v", i)
	}

	i, err = st.Next()
	if err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatalf("expected idx 1, got %v", i)
	}
	i, err = st.Next()
	if err != nil {
		t.Fatal(err)
	}
	if i != 2 {
		t.Fatalf("expected idx 2, got %v", i)
	}
	if err != nil {
		t.Fatal(err)
	}

	s, i = st.Where()
	if s != "bar" {
		t.Fatalf("expected baz, got %s", s)
	}
	if i != 2 {
		t.Fatalf("expected idx 2, got %v", i)
	}

	s, err = st.Up()
	if err != nil {
		t.Fatal(err)
	}
	if s != "foo" {
		t.Fatalf("expected foo, got %s", s)
	}
	s, i = st.Where()
	if s != "foo" {
		t.Fatalf("expected foo, got %s", s)
	}
	if i != 0 {
		t.Fatalf("expected idx 0, got %v", i)
	}
}

func TestStateFlagMatch(t *testing.T) {
	st := NewState(2)
	st.SetFlag(8)
	v := st.MatchFlag(8, true)
	if !v {
		t.Fatalf("unexpected flag")
	}
	v = st.MatchFlag(8, false)
	if v {
		t.Fatalf("unexpected flag")
	}

	v = st.MatchFlag(9, true)
	if v {
		t.Fatalf("unexpected flag")
	}
	v = st.MatchFlag(9, false)
	if !v {
		t.Fatalf("unexpected flag")
	}
}

func TestStateMovementNoRoot(t *testing.T) {
	var err error
	st := NewState(0)
	_, err = st.Next()
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = st.Previous()
	if err == nil {
		t.Fatal("expected error")
	}
	vl, vr := st.Sides()
	if vl {
		t.Fatal("expected false")
	}
	if vr {
		t.Fatal("expected false")
	}
	_, err = st.Top()
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = st.Up()
	if err == nil {
		t.Fatal("expected error")
	}
	v := st.Depth()
	if v != -1 {
		t.Fatalf("expected -1, got %d", v)
	}
}

func TestStateInput(t *testing.T) {
	var err error
	var wrongInput [257]byte
	st := NewState(0)
	_, err = st.GetInput()
	if err == nil {
		t.Fatal("expected error")
	}
	err = st.SetInput(wrongInput[:])
	if err == nil {
		t.Fatal("expected error")
	}
	b := []byte("foo")
	err = st.SetInput(b)
	if err != nil {
		t.Fatal(err)
	}
	v, err := st.GetInput()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, b) {
		t.Fatalf("expected %x, got %x", b, v)
	}
}

func TestStateMovement(t *testing.T) {
	var x uint16
	st := NewState(0)
	err := st.Down("foo")
	if err != nil {
		t.Fatal(err)
	}
	r, err := st.Top()
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Fatal("expected true")
	}
	err = st.Down("bar")
	if err != nil {
		t.Fatal(err)
	}
	err = st.Down("baz")
	if err != nil {
		t.Fatal(err)
	}
	v := st.Depth()
	if v != 2 {
		t.Fatalf("expected 1, got %d", v)
	}
	s, err := st.Up()
	if err != nil {
		t.Fatal(err)
	}
	if s != "bar" {
		t.Fatalf("expected 'bar', got '%s'", s)
	}
	v = st.Depth()
	if v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}
	vr, vl := st.Sides()
	if !vr {
		t.Fatal("expected true")
	}
	if vl {
		t.Fatal("expected false")
	}
	_, err = st.Previous()
	if err == nil {
		t.Fatal("expected error")
	}
	x, err = st.Next()
	if err != nil {
		t.Fatal(err)
	}
	if x != 1 {
		t.Fatalf("expected 1, got %d", x)
	}
	vr, vl = st.Sides()
	if !vr {
		t.Fatal("expected true")
	}
	if !vl {
		t.Fatal("expected true")
	}
	x, err = st.Next()
	if err != nil {
		t.Fatal(err)
	}
	if x != 2 {
		t.Fatalf("expected 2, got %d", x)
	}
	_, err = st.Next()
	if err != nil {
		t.Fatal(err)
	}
	s, x = st.Where()
	if s != "bar" {
		t.Fatalf("expected 'baz', got '%s'", s)
	}
	if x != 3 {
		t.Fatalf("expected 3, got '%d'", x)
	}
	x, err = st.Previous()
	if x != 2 {
		t.Fatalf("expected 2, got '%d'", x)
	}
	vl, vr = st.Sides()
	if !vr {
		t.Fatal("expected true")
	}
	if !vl {
		t.Fatal("expected true")
	}
}

func TestStateMaxMovement(t *testing.T) {
	MaxLevel = 3
	st := NewState(0)
	st.Down("inky")	
	st.Down("pinky")	
	st.Down("blinky")
	st.Down("clyde")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	st.Down("sue")
}

func TestStateReset(t *testing.T) {
	st := NewState(3)
	st.SetFlag(2)
	st.SetFlag(9)
	st.Down("foo")
	st.Down("bar")
	st.Down("baz")
	st.Next()
	st.Next()
	st.SetInput([]byte("xyzzy"))
	err := st.Restart()
	if err != nil {
		t.Fatal(err)
	}
	r, err := st.Top()
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Fatal("expected true")
	}
	if st.GetFlag(2) {
		t.Fatal("expected not set")
	}
	s, v := st.Where()
	if s != "foo" {
		t.Fatalf("expected 'foo', got '%s'", s)
	}
	if v > 0 {
		t.Fatalf("expected 0, got %d", v)
	}
	
}

func TestStateLanguage(t *testing.T) {
	st := NewState(0)
	if st.Language != nil {
		t.Fatal("expected language not set")
	}
	err := st.SetLanguage("nor")
	if err != nil {
		t.Fatal(err)
	}
	if st.Language == nil {
		t.Fatal("expected language set")
	}
}
