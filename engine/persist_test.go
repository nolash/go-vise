package engine

import (
	"context"
	"os"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/db"
)

func TestRunPersist(t *testing.T) {
	generateTestData(t)
	ctx := context.Background()
	cfg := Config{
		OutputSize: 83,
		SessionId: "xyzzy",
		Root: "root",
	}
	rs := newTestWrapper(dataDir, nil)

	st := state.NewState(3)
	ca := cache.NewCache().WithCacheSize(1024)
	store := db.NewMemDb()
	store.Connect(ctx, "")
	pr := persist.NewPersister(store).WithContent(&st, ca)

	w := os.Stdout
	ctx = context.Background()

	st = state.NewState(cfg.FlagCount)
	ca = cache.NewCache()
	ca = ca.WithCacheSize(cfg.CacheSize)
	pr = persist.NewPersister(store).WithContent(&st, ca)
	err := pr.Save(cfg.SessionId)
	if err != nil {
		t.Fatal(err)
	}
	
	pr = persist.NewPersister(store)
	inputs := []string{
		"", // trigger init, will not exec
		"1",
		"2",
		"00",
		}
	for _, v := range inputs {
		err = RunPersisted(cfg, rs, pr, []byte(v), w, ctx)
		if err != nil {
			t.Fatal(err)
		}
	}

	pr = persist.NewPersister(store)
	err = pr.Load(cfg.SessionId)
	if err != nil {
		t.Fatal(err)
	}

	stAfter := pr.GetState()
	location, idx := stAfter.Where()
	if location != "long" {
		t.Fatalf("expected 'long', got %s", location)
	}
	if idx != 1 {
		t.Fatalf("expected '1', got %v", idx)
	}
}

func TestEnginePersist(t *testing.T) {
	generateTestData(t)
	ctx := context.Background()
	cfg := Config{
		OutputSize: 83,
		SessionId: "xyzzy",
		Root: "root",
	}
	rs := newTestWrapper(dataDir, nil)

	st := state.NewState(3)
	ca := cache.NewCache().WithCacheSize(1024)
	store := db.NewMemDb()
	store.Connect(ctx, "")
	pr := persist.NewPersister(store).WithContent(&st, ca)

	st = state.NewState(cfg.FlagCount)
	ca = cache.NewCache()
	ca = ca.WithCacheSize(cfg.CacheSize)
	pr = persist.NewPersister(store).WithContent(&st, ca)
	err := pr.Save(cfg.SessionId)
	if err != nil {
		t.Fatal(err)
	}

	en, err := NewPersistedEngine(ctx, cfg, pr, rs)
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.Exec(ctx, []byte("1"))
	if err != nil {
		t.Fatal(err)
	}
	
	_, err = en.Exec(ctx, []byte("2"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = en.Exec(ctx, []byte("00"))
	if err != nil {
		t.Fatal(err)
	}
	location, idx := st.Where()
	if location != "long" {
		t.Fatalf("expected location 'long', got %s", location)
	}
	if idx != 1 {
		t.Fatalf("expected index '1', got %v", idx)
	}

	pr = persist.NewPersister(store)
	en, err = NewPersistedEngine(ctx, cfg, pr, rs)
	if err != nil {
		t.Fatal(err)
	}
	st_loaded := pr.GetState()
	location, _ = st_loaded.Where()
	if location != "long" {
		t.Fatalf("expected location 'long', got %s", location)
	}
	if idx != 1 {
		t.Fatalf("expected index '1', got %v", idx)
	}

	_, err = en.Exec(ctx, []byte("11"))
	if err != nil {
		t.Fatal(err)
	}
}
