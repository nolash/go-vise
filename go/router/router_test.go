package router

import (
	"bytes"
	"testing"
)

func TestRouter(t *testing.T) {
	r := NewRouter()
	err := r.Add("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	err = r.Add("baz", "barbarbar")
	if err != nil {
		t.Error(err)
	}
	err = r.Add("foo", "xyzzy")
	if err == nil {
		t.Errorf("expected error for duplicate key foo")
	}
}

func TestRouterOut(t *testing.T) {
	rt := NewRouter()
	err := rt.Add("foo", "inky")
	if err != nil {
		t.Error(err)
	}
	err = rt.Add("barbar", "pinky")
	if err != nil {
		t.Error(err)
	}
	err = rt.Add("bazbazbaz", "blinky")
	if err != nil {
		t.Error(err)
	}
	rb := []byte{}
	r := rt.Next()
	expect := append([]byte{0x3}, []byte("foo")...)
	expect = append(expect, 4)
	expect = append(expect, []byte("inky")...)
	if !bytes.Equal(r, expect) {
		t.Errorf("expected %v, got %v", expect, r)
	}
	rb = append(rb, r...)

	r = rt.Next()
	expect = append([]byte{0x6}, []byte("barbar")...)
	expect = append(expect, 5)
	expect = append(expect, []byte("pinky")...)
	if !bytes.Equal(r, expect) {
		t.Errorf("expected %v, got %v", expect, r)
	}
	rb = append(rb, r...)

	r = rt.Next()
	expect = append([]byte{0x9}, []byte("bazbazbaz")...)
	expect = append(expect, 6)
	expect = append(expect, []byte("blinky")...)
	if !bytes.Equal(r, expect) {
		t.Errorf("expected %v, got %v", expect, r)
	}
	rb = append(rb, r...)
}

func TestSerialize(t *testing.T) {
	rt := NewRouter()
	err := rt.Add("foo", "inky")
	if err != nil {
		t.Error(err)
	}
	err = rt.Add("barbar", "pinky")
	if err != nil {
		t.Error(err)
	}
	err = rt.Add("bazbazbaz", "blinky")
	if err != nil {
		t.Error(err)
	}

	// Serialize and deserialize.
	ra := rt.ToBytes()
	rt = FromBytes(ra)
	rb := rt.ToBytes()
	if !bytes.Equal(ra, rb) {
		t.Errorf("expected %v, got %v", ra, rb)
	}
}
