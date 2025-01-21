package engine

import (
	"bytes"
	"context"
	"testing"

	"git.defalsify.org/vise.git/cache"
	memdb "git.defalsify.org/vise.git/db/mem"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/state"
)

func TestPersistNewAcrossEngine(t *testing.T) {
	var err error
	var cfg Config
	generateTestData(t)
	st := state.NewState(1)
	ca := cache.NewCache()
	rs := newTestWrapper(dataDir, st)
	ctx := context.Background()
	store := memdb.NewMemDb()
	store.Connect(ctx, "")
	pe := persist.NewPersister(store)
	en := NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	en = en.WithPersister(pe)
	cont, err := en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Errorf("expected cont")
	}

	r := bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, r)
	if err != nil {
		t.Fatal(err)
	}

	cont, err = en.Exec(ctx, []byte("1"))
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Errorf("expected cont")
	}
	r = bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, r)
	if err != nil {
		t.Fatal(err)
	}

	err = en.Finish(ctx)
	if err != nil {
		t.Fatal(err)
	}

	cfg.FlagCount = 1
	pe = persist.NewPersister(store)
	en = NewEngine(cfg, rs)
	en = en.WithPersister(pe)
	cont, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Errorf("expected cont")
	}
	location, _ := st.Where()
	if location != "foo" {
		t.Errorf("expected location 'foo', got '%s", location)
	}
}

func TestPersistSameAcrossEngine(t *testing.T) {
	var err error
	var cfg Config
	generateTestData(t)
	st := state.NewState(1)
	ca := cache.NewCache()
	rs := newTestWrapper(dataDir, st)
	ctx := context.Background()
	store := memdb.NewMemDb()
	store.Connect(ctx, "")
	pe := persist.NewPersister(store)
	pe = pe.WithFlush()
	en := NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	en = en.WithPersister(pe)
	cont, err := en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Errorf("expected cont")
	}

	r := bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, r)
	if err != nil {
		t.Fatal(err)
	}

	cont, err = en.Exec(ctx, []byte("1"))
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Errorf("expected cont")
	}
	r = bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, r)
	if err != nil {
		t.Fatal(err)
	}

	err = en.Finish(ctx)
	if err != nil {
		t.Fatal(err)
	}

	cfg.FlagCount = 1
	en = NewEngine(cfg, rs)
	en = en.WithPersister(pe)
	cont, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Errorf("expected cont")
	}
	location, _ := st.Where()
	if location != "foo" {
		t.Errorf("expected location 'foo', got '%s", location)
	}
}
