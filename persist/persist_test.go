package persist

import (
	"context"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/cache"
)

func TestInvalidateState(t *testing.T) {
	st := state.NewState(0)
	ca := cache.NewCache()

	ctx := context.Background()
	store := db.NewMemDb(ctx)
	store.Connect(ctx, "")
	pr := NewPersister(store).WithSession("xyzzy").WithContent(&st, ca)
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
	store := db.NewMemDb(ctx)
	store.Connect(ctx, "")
	pr := NewPersister(store).WithSession("xyzzy").WithContent(&st, ca)
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
	store := db.NewMemDb(ctx)
	store.Connect(ctx, "")
	pr := NewPersister(store).WithSession("xyzzy").WithContent(&st, ca)
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
