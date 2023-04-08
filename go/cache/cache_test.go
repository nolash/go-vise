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
	err = ca.Pop()
	if err != nil {
		t.Error(err)
	}
	err = ca.Pop()
	if err == nil {
		t.Errorf("expected out of top frame error")
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

func TestCacheCurrentSize(t *testing.T) {
	ca := NewCache()
	err := ca.Push()
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("foo", "inky", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Push()
	err = ca.Add("bar", "pinky", 10)
	if err != nil {
		t.Error(err)
	}
	err = ca.Map("bar")
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("baz", "tinkywinkydipsylalapoo", 51)
	if err != nil {
		t.Error(err)
	}
	err = ca.Map("baz")
	if err != nil {
		t.Error(err)
	}

	l, c := ca.Usage()
	if l != 27 {
		t.Errorf("expected actual length 27, got %v", l)
	}
	if c != 34 {
		t.Errorf("expected remaining length 34, got %v", c)
	}
}

func TestStateMapSink(t *testing.T) {
	ca := NewCache()
	ca.Push()
	err := ca.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	ca.Push()
	err = ca.Add("bar", "xyzzy", 6)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("baz", "bazbaz", 18)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("xyzzy", "plugh", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Map("foo")
	if err != nil {
		t.Error(err)
	}
	err = ca.Map("xyzzy")
	if err == nil {
		t.Errorf("Expected fail on duplicate sink")
	}
	err = ca.Map("baz")
	if err != nil {
		t.Error(err)
	}
	ca.Push()
	err = ca.Map("foo")
	if err != nil {
		t.Error(err)
	}
	ca.Pop()
	err = ca.Map("foo")
	if err != nil {
		t.Error(err)
	}
}
