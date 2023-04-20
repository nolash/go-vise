package engine

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/state"
)

func TestPersist(t *testing.T) {
	generateTestData(t)
	cfg := Config{
		OutputSize: 83,
		SessionId: "xyzzy",
		Root: "root",
	}
	rs := NewFsWrapper(dataDir, nil)

	persistDir, err := ioutil.TempDir("", "vise_engine_persist")
	if err != nil {
		t.Fatal(err)
	}

	st := state.NewState(3)
	ca := cache.NewCache().WithCacheSize(1024)
	pr := persist.NewFsPersister(persistDir).WithContent(&st, ca)

	w := os.Stdout
	ctx := context.TODO()

	st = state.NewState(cfg.FlagCount)
	ca = cache.NewCache()
	ca = ca.WithCacheSize(cfg.CacheSize)
	pr = persist.NewFsPersister(persistDir).WithContent(&st, ca)
	err = pr.Save(cfg.SessionId)
	if err != nil {
		t.Fatal(err)
	}
	
	pr = persist.NewFsPersister(persistDir)
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

	pr = persist.NewFsPersister(persistDir)
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
