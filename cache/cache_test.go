package cache

import (
	"slices"
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

func TestCacheUse(t *testing.T) {
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
	v := ca.Check("inky")
	if v {
		t.Fatal("expected true")
	}
	v = ca.Check("blinky")
	if !v {
		t.Fatal("expected false")
	}
}

func TestCacheUpdate(t *testing.T) {
	ca := NewCache()
	ca = ca.WithCacheSize(10)
	ca.Add("foo", "bar", 0)
	err := ca.Add("foo", "barbarbar", 0)
	if err != ErrDup {
		t.Error(err)
	}
	v, err := ca.Get("foo")
	if err != nil {
		t.Error(err)
	}
	if v != "bar" {
		t.Fatalf("expected 'bar', got '%s'", v)
	}
	err = ca.Update("foo", "barbarbar")
	v, err = ca.Get("foo")
	if err != nil {
		t.Error(err)
	}
	if v != "barbarbar" {
		t.Fatalf("expected 'barbarbar', got '%s'", v)
	}
	err = ca.Update("foo", "barbarbarbar")
	if err == nil {
		t.Fatalf("expect error")
	}
}

func TestCacheLimits(t *testing.T) {
	ca := NewCache()
	ca = ca.WithCacheSize(8)
	err := ca.Add("foo", "bar", 2)
	if err == nil {
		t.Fatal("expected error")
	}
	err = ca.Add("foo", "barbarbar", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	err = ca.Add("foo", "bar", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = ca.Add("baz", "barbar", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	ca.Reset()
	err = ca.Add("baz", "barbar", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	ca.Pop()
	err = ca.Add("baz", "barbar", 0)
	if err != nil {
		t.Fatal(err)
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
	ca.Push()
	err = ca.Add("baz", "xyzzy", 0)
	if err != nil {
		t.Error(err)
	}
	ca.Reset()
	if ca.CacheUseSize != 3 {
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
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("foo", "baz", 0)
	if err == nil {
		t.Errorf("expect duplicate key in different frame")
	}
	ca.Pop()
	err = ca.Add("foo", "baz", 0)
	if err != ErrDup {
		t.Error(err)
	}
}

func TestCacheLast(t *testing.T) {
	ca := NewCache()
	v := ca.Last()
	if v != "" {
		t.Fatal("expected empty")
	}
	err := ca.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	ca.Push()
	err = ca.Add("baz", "xyzzy", 0)
	if err != nil {
		t.Error(err)
	}
	v = ca.Last()
	if v != "xyzzy" {
		t.Fatalf("expected 'xyzzy', got: '%s'", v)
	}
}

func TestCacheKeys(t *testing.T) {
	ca := NewCache()
	ca.Add("inky", "tinkywinky", 0)
	ca.Push()
	ca.Add("pinky", "dipsy", 0)
	ca.Push()
	ca.Push()
	ca.Add("blinky", "lala", 0)
	ca.Add("clyde", "pu", 0)
	ks := ca.Keys(3)
	if !slices.Contains(ks, "blinky") {
		t.Fatalf("Missing 'blinky'")
	}
	if !slices.Contains(ks, "clyde") {
		t.Fatalf("Missing 'clyde'")
	}
}
