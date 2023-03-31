package state

import (
	"testing"
)

func TestNewStateFlags(t *testing.T) {
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

func TestStateEnterExit(t *testing.T) {
	st := NewState(17)
	st.Enter("one")
	err := st.Add("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	err = st.Add("baz", "xyzzy")
	if err != nil {
		t.Error(err)
	}
	if st.CacheUseSize != 8 {
		t.Errorf("expected cache use size 8 got %v", st.CacheUseSize)
	}
	err = st.Exit()
	if err != nil {
		t.Error(err)
	}
	err = st.Exit()
	if err == nil {
		t.Errorf("expected out of top frame error")
	}
}

func TestStateReset(t *testing.T) {
	st := NewState(17)
	st.Enter("one")
	err := st.Add("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	err = st.Add("baz", "xyzzy")
	if err != nil {
		t.Error(err)
	}
	st.Enter("two")
	st.Enter("three")
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
	st.Enter("one")
	err := st.Add("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	st.Enter("two")
	err = st.Add("foo", "baz")
	if err == nil {
		t.Errorf("expected fail on duplicate load")
	}
}
