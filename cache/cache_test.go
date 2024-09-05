package cache

import (
	"testing"
)

func TestNewCache(t *testing.T) {
	ca := NewCache()
	if ca.CacheSize != 0 {
		t.Errorf("cache size not 0")
	}
	ca = ca.WithCacheSize(102525)
	if ca.CacheSize != 102525 {
		t.Errorf("cache size not 102525")
	}
}

func TestStateCacheUse(t *testing.T) {
	ca := NewCache()
	ca = ca.WithCacheSize(10)
	ca.Push()
	err := ca.Add("bar", "baz", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("inky", "pinky", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("blinky", "clyde", 0)
	if err == nil {
		t.Errorf("expected capacity error")
	}
}

func TestStateDownUp(t *testing.T) {
	ca := NewCache()
	err := ca.Push()
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("baz", "xyzzy", 0)
	if err != nil {
		t.Error(err)
	}
	if ca.CacheUseSize != 8 {
		t.Errorf("expected cache use size 8 got %v", ca.CacheUseSize)
	}
	err = ca.Pop()
	if err != nil {
		t.Error(err)
	}
	l := len(ca.Cache)
	if l != 1 {
		t.Fatalf("expected cache length 1, got %d", l)
	}
	err = ca.Pop()
	if err != nil {
		t.Error(err)
	}
	l = len(ca.Cache)
	if l != 1 {
		t.Fatalf("expected cache length 1, got %d", l)
	}
	err = ca.Pop()
	if err != nil {
		t.Errorf("unexpected out of top frame error")
	}
	l = len(ca.Cache)
	if l != 1 {
		t.Fatalf("expected cache length 1, got %d", l) 
	}
}

func TestCacheReset(t *testing.T) {
	ca := NewCache()
	err := ca.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("baz", "xyzzy", 0)
	if err != nil {
		t.Error(err)
	}
	ca.Reset()
	if ca.CacheUseSize != 0 {
		t.Errorf("expected cache use size 0, got %v", ca.CacheUseSize)
	}
}

func TestCacheLoadDup(t *testing.T) {
	ca := NewCache()
	err := ca.Push()
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("foo", "xyzzy", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Push()
	err = ca.Add("foo", "baz", 0)
	if err == nil {
		t.Errorf("expected fail on duplicate load")
	}
	ca.Pop()
	err = ca.Add("foo", "baz", 0)
	if err != nil {
		t.Error(err)
	}
}

