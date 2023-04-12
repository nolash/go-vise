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
}

func TestStateflags(t *testing.T) {
	st := NewState(9)
	v, err := st.GetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if v {
		t.Fatalf("Expected bit 2 not to be set")
	}
	v, err = st.SetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Fatalf("Expected change to be set for bit 2")
	}
	v, err = st.GetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Fatalf("Expected bit 2 to be set")
	}
	v, err = st.SetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Fatalf("Expected change to be set for bit 10")
	}
	v, err = st.GetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Fatalf("Expected bit 10 to be set")
	}
	v, err = st.ResetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Fatalf("Expected change to be set for bit 10")
	}
	v, err = st.GetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if v {
		t.Fatalf("Expected bit 2 not to be set")
	}
	v, err = st.GetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Fatalf("Expected bit 10 to be set")
	}
	v, err = st.SetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if v {
		t.Fatalf("Expected change not to be set for bit 10")
	}
	v, err = st.SetFlag(2)
	if err != nil {
		t.Error(err)
	}
	v, err = st.SetFlag(16)
	if err != nil {
		t.Error(err)
	}
	v, err = st.SetFlag(17)
	if err == nil {
		t.Fatalf("Expected out of range for bit index 17")
	}
	if !bytes.Equal(st.Flags[:3], []byte{0x04, 0x04, 0x01}) {
		t.Fatalf("Expected 0x040401, got %v", st.Flags[:3])
	}
}

func TestStateFlagFromSlice(t *testing.T) {
	st := NewState(15)
	_, _= st.SetFlag(2)
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
	_, _= st.SetFlag(12)
	v = st.GetIndex([]byte{0x04})
	if !v {
		t.Fatalf("Expected 0x04 to match")
	}
	v = st.GetIndex([]byte{0x00, 0x10})
	if !v {
		t.Fatalf("Expected 0x1000 to match")
	}
	v, _ = st.ResetFlag(2)
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
