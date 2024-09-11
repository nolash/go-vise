package persist

import (
	"context"
	"testing"

	"git.defalsify.org/vise.git/db/mem"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/cache"
)

func TestCreateCache(t *testing.T) {
	ca := cache.NewCache()
	if ca.Levels() != 1 {
		t.Fatalf("expected level 1, got: %d", ca.Levels())
	}
	ca.Push()
	ca.Push()
	if ca.Levels() != 3 {
		t.Fatalf("expected level 3, got: %d", ca.Levels())
	}
	ca.Reset()
	if ca.Levels() != 1 {
		t.Fatalf("expected level 1, got: %d", ca.Levels())
	}
}

func TestCacheUseSize(t *testing.T) {
	ca := cache.NewCache()
	v := ca.CacheUseSize
	if v != 0 {
		t.Fatalf("expected cache use size 0, got: %v", v)
	}
	ca.Add("foo", "barbarbar", 12)
	v = ca.CacheUseSize
	if v != 9 {
		t.Fatalf("expected cache use size 9, got: %v", v)
	}
	ca.Reset()
	v = ca.CacheUseSize
	if v != 9 {
		t.Fatalf("expected cache use size 9, got: %v", v)
	}
	ca.Pop()
	v = ca.CacheUseSize
	if v != 0 {
		t.Fatalf("expected cache use size 0, got: %v", v)
	}
}

func TestInvalidateState(t *testing.T) {
	st := state.NewState(0)
	ca := cache.NewCache()

	ctx := context.Background()
	store := mem.NewMemDb()
	store.Connect(ctx, "")
	pr := NewPersister(store).WithSession("xyzzy").WithContent(st, ca)
	err := pr.Save("foo")
	if err != nil {
		t.Fatal(err)
	}

	st.Invalidate()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")	
		}
	}()
	_ = pr.Save("foo")
}

func TestInvalidateCache(t *testing.T) {
	st := state.NewState(0)
	ca := cache.NewCache()

	ctx := context.Background()
	store := mem.NewMemDb()
	store.Connect(ctx, "")
	pr := NewPersister(store).WithSession("xyzzy").WithContent(st, ca)
	err := pr.Save("foo")
	if err != nil {
		t.Fatal(err)
	}

	ca.Invalidate()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")	
		}
	}()
	_ = pr.Save("foo")
}

func TestInvalidateAll(t *testing.T) {
	st := state.NewState(0)
	ca := cache.NewCache()

	ctx := context.Background()
	store := mem.NewMemDb()
	store.Connect(ctx, "")
	pr := NewPersister(store).WithSession("xyzzy").WithContent(st, ca)
	err := pr.Save("foo")
	if err != nil {
		t.Fatal(err)
	}

	ca.Invalidate()
	st.Invalidate()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")	
		}
	}()
	_ = pr.Save("foo")
}

func TestFlush(t *testing.T) {
	ctx := context.Background()
	st := state.NewState(15)
	ca := cache.NewCache()
	store := mem.NewMemDb()
	store.Connect(ctx, "")

	ca.Add("foo", "bar", 0)
	ca.Push()
	ca.Add("inky", "pinky", 42)
	ca.Push()
	ca.Add("blinky", "clyde", 13)
	ca.WithCacheSize(666)
	
	st.Down("xyzzy")
	st.Down("plugh")
	st.SetFlag(3)
	st.SetFlag(10)
	st.SetFlag(19)

	pe := NewPersister(store).WithContent(st, ca).WithFlush()
	err := pe.Save("baz")
	if err != nil {
		t.Fatal(err)
	}
	expectBitSize := uint32(15 + 8)
	if st.FlagBitSize() != expectBitSize {
		t.Fatalf("expected bitsize %d, got %d", expectBitSize, st.FlagBitSize())
	}
	st = pe.GetState()
	node,  lvl := st.Where()
	if lvl != 0 {
		t.Fatalf("expected level 0, got: %d", lvl)
	}
	if node != "" {
		t.Fatalf("expected node '', got '%s'", node)
	}
	cm := pe.GetMemory()
	if cm.Levels() != 1 {
		t.Fatalf("expected level 1, got: %d", cm.Levels())
	}
	_, err = cm.Get("foo")
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = cm.Get("blinky")
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = cm.ReservedSize("foo")
	if err == nil {
		t.Fatal("expected error")
	}
	ks := cm.Keys(0)
	if len(ks) > 0 {
		t.Fatalf("expected keys list length 0, got: %v", ks)
	}
	o, ok := cm.(*cache.Cache)
	if !ok {
		panic("not cache")
	}
	if o.CacheUseSize != 0 {
		t.Fatalf("expected cache use size 0, got: %v", o.CacheUseSize)
	}
}
