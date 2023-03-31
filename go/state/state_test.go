package state

import (
	"bytes"
	"testing"
)

// Check creation 
func TestNewState(t *testing.T) {
	st := NewState(5)
	if len(st.Flags) != 1 {
		t.Errorf("invalid state flag length: %v", len(st.Flags))
	}
	st = NewState(8)
	if len(st.Flags) != 1 {
		t.Errorf("invalid state flag length: %v", len(st.Flags))
	}
	st = NewState(17)
	if len(st.Flags) != 3 {
		t.Errorf("invalid state flag length: %v", len(st.Flags))
	}
}

func TestStateFlags(t *testing.T) {
	st := NewState(17)
	v, err := st.GetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if v {
		t.Errorf("Expected bit 2 not to be set")
	}
	v, err = st.SetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Errorf("Expected change to be set for bit 2")
	}
	v, err = st.GetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Errorf("Expected bit 2 to be set")
	}
	v, err = st.SetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Errorf("Expected change to be set for bit 10")
	}
	v, err = st.GetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Errorf("Expected bit 10 to be set")
	}
	v, err = st.ResetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Errorf("Expected change to be set for bit 10")
	}
	v, err = st.GetFlag(2)
	if err != nil {
		t.Error(err)
	}
	if v {
		t.Errorf("Expected bit 2 not to be set")
	}
	v, err = st.GetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if !v {
		t.Errorf("Expected bit 10 to be set")
	}
	v, err = st.SetFlag(10)
	if err != nil {
		t.Error(err)
	}
	if v {
		t.Errorf("Expected change not to be set for bit 10")
	}
	v, err = st.SetFlag(2)
	if err != nil {
		t.Error(err)
	}
	v, err = st.SetFlag(19)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(st.Flags[:3], []byte{0x04, 0x04, 0x08}) {
		t.Errorf("Expected 0x020203, got %v", st.Flags[:3])
	}
}

//
func TestNewStateCache(t *testing.T) {
	st := NewState(17)
	if st.CacheSize != 0 {
		t.Errorf("cache size not 0")
	}
	st = st.WithCacheSize(102525)
	if st.CacheSize != 102525 {
		t.Errorf("cache size not 102525")
	}

}

func TestStateCacheUse(t *testing.T) {
	st := NewState(17)
	st = st.WithCacheSize(10)
	st.Down("foo")
	err := st.Add("bar", "baz", 0)
	if err != nil {
		t.Error(err)
	}
	err = st.Add("inky", "pinky", 0)
	if err != nil {
		t.Error(err)
	}
	err = st.Add("blinky", "clyde", 0)
	if err == nil {
		t.Errorf("expected capacity error")
	}
}

func TestStateDownUp(t *testing.T) {
	st := NewState(17)
	st.Down("one")
	err := st.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	err = st.Add("baz", "xyzzy", 0)
	if err != nil {
		t.Error(err)
	}
	if st.CacheUseSize != 8 {
		t.Errorf("expected cache use size 8 got %v", st.CacheUseSize)
	}
	err = st.Up()
	if err != nil {
		t.Error(err)
	}
	err = st.Up()
	if err != nil {
		t.Error(err)
	}
	err = st.Up()
	if err == nil {
		t.Errorf("expected out of top frame error")
	}
}

func TestStateReset(t *testing.T) {
	st := NewState(17)
	st.Down("one")
	err := st.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	err = st.Add("baz", "xyzzy", 0)
	if err != nil {
		t.Error(err)
	}
	st.Down("two")
	st.Down("three")
	st.Reset()
	if st.CacheUseSize != 0 {
		t.Errorf("expected cache use size 0, got %v", st.CacheUseSize)
	}
	if st.Depth() != 1 {
		t.Errorf("expected depth 1, got %v", st.Depth())
	}
}

func TestStateLoadDup(t *testing.T) {
	st := NewState(17)
	st.Down("one")
	err := st.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	st.Down("two")
	err = st.Add("foo", "baz", 0)
	if err == nil {
		t.Errorf("expected fail on duplicate load")
	}
}

func TestStateCurrentSize(t *testing.T) {
	st := NewState(17)
	st.Down("one")
	err := st.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	st.Down("two")
	err = st.Add("bar", "xyzzy", 10)
	if err != nil {
		t.Error(err)
	}
	err = st.Map("bar")
	if err != nil {
		t.Error(err)
	}
	err = st.Add("baz", "inkypinkyblinkyclyde", 51)
	if err != nil {
		t.Error(err)
	}
	err = st.Map("baz")
	if err != nil {
		t.Error(err)
	}
	l, c := st.Size()
	if l != 25 {
		t.Errorf("expected actual length 25, got %v", l)
	}
	if c != 36 {
		t.Errorf("expected actual length 50, got %v", c)
	}
}

func TestStateMapSink(t *testing.T) {
	st := NewState(17)
	st.Down("one")
	err := st.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	st.Down("two")
	err = st.Add("bar", "xyzzy", 6)
	if err != nil {
		t.Error(err)
	}
	err = st.Add("baz", "bazbaz", 18)
	if err != nil {
		t.Error(err)
	}
	err = st.Add("xyzzy", "plugh", 0)
	if err != nil {
		t.Error(err)
	}
	err = st.Map("foo")
	if err != nil {
		t.Error(err)
	}
	err = st.Map("xyzzy")
	if err == nil {
		t.Errorf("Expected fail on duplicate sink")
	}
	err = st.Map("baz")
	if err != nil {
		t.Error(err)
	}
	st.Down("three")
	err = st.Map("foo")
	if err != nil {
		t.Error(err)
	}
	st.Up()
	err = st.Map("foo")
	if err != nil {
		t.Error(err)
	}
}
