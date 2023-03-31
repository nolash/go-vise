package state

import (
	"testing"
)

func TestNewStateFlags(t *testing.T) {
	st := NewState(5, 0)
	if len(st.Flags) != 1 {
		t.Errorf("invalid state flag length: %v", len(st.Flags))
	}
	st = NewState(8, 0)
	if len(st.Flags) != 1 {
		t.Errorf("invalid state flag length: %v", len(st.Flags))
	}
	st = NewState(17, 0)
	if len(st.Flags) != 3 {
		t.Errorf("invalid state flag length: %v", len(st.Flags))
	
	}
}

func TestNewStateCache(t *testing.T) {
	st := NewState(17, 0)
	if st.CacheSize != 0 {
		t.Errorf("cache size not 0")
	}
	st = st.WithCacheSize(102525)
	if st.CacheSize != 102525 {
		t.Errorf("cache size not 102525")
	}

}

func TestStateCacheUse(t *testing.T) {
	st := NewState(17, 0)
	st = st.WithCacheSize(10)
	st.Enter("foo")
	err := st.Add("bar", "baz")
	if err != nil {
		t.Error(err)
	}
	err = st.Add("inky", "pinky")
	if err != nil {
		t.Error(err)
	}
	err = st.Add("blinky", "clyde")
	if err == nil {
		t.Errorf("expected capacity error")
	}
}
