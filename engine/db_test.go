package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/vm"
	memdb "git.defalsify.org/vise.git/db/mem"
)

func getNull() io.WriteCloser {
	nul, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0700)
	if err != nil {
		panic(err)
	}
	return nul
}

func codeGet(ctx context.Context, s string) ([]byte, error) {
	var b []byte
	var err error
	switch s {
		case "root":
			b = vm.NewLine(nil, vm.HALT, nil, nil, nil)
			b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{0x0}, nil)
		case "tinkywinky":
			b = vm.NewLine(nil, vm.MOVE, []string{"dipsy"}, nil, nil)
			b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{0x0}, nil)
		case "dipsy":
			b = vm.NewLine(nil, vm.HALT, nil, nil, nil)
			b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{0x0}, nil)
		default:
			err = fmt.Errorf("unknown code symbol '%s'", s)
	}
	return b, err
}

func flagSet(ctx context.Context, nodeSym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "xyzzy",
		FlagSet: []uint32{state.FLAG_USERSTART},
	}, nil
}

func TestDbEngineMinimal(t *testing.T) {
	ctx := context.Background()
	cfg := Config{}
	rs := resource.NewMenuResource()
	en := NewEngine(cfg, rs)
	cont, err := en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cont {
		t.Fatalf("expected not continue")
	}
	err = en.Finish()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDbEngineNoResource(t *testing.T) {
	cfg := Config{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	NewEngine(cfg, nil)
}

func TestDbEngineStateNil(t *testing.T) {
	cfg := Config{}
	rs := resource.NewMenuResource()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	NewEngine(cfg, rs).WithState(nil)
}

func TestDbEngineCacheNil(t *testing.T) {
	cfg := Config{}
	rs := resource.NewMenuResource()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	NewEngine(cfg, rs).WithMemory(nil)
}

func TestDbEnginePersisterNil(t *testing.T) {
	cfg := Config{}
	rs := resource.NewMenuResource()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	NewEngine(cfg, rs).WithPersister(nil)
}

func TestDbEngineFirstNil(t *testing.T) {
	cfg := Config{}
	rs := resource.NewMenuResource()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	NewEngine(cfg, rs).WithFirst(nil)
}

func TestDbEngineStateDup(t *testing.T) {
	cfg := Config{}
	rs := resource.NewMenuResource()
	st := state.NewState(0)
	en := NewEngine(cfg, rs).WithState(st)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	en.WithState(st)
}


func TestDbEngineCacheDup(t *testing.T) {
	cfg := Config{}
	rs := resource.NewMenuResource()
	ca := cache.NewCache()
	en := NewEngine(cfg, rs).WithMemory(ca)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	en.WithMemory(ca)
}

func TestDbEnginePersisterDup(t *testing.T) {
	ctx := context.Background()
	cfg := Config{}
	rs := resource.NewMenuResource()
	store := memdb.NewMemDb()
	store.Connect(ctx, "")
	pe := persist.NewPersister(store)
	en := NewEngine(cfg, rs).WithPersister(pe)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	en.WithPersister(pe)
}

func TestDbEngineFirstDup(t *testing.T) {
	cfg := Config{}
	rs := resource.NewMenuResource()
	en := NewEngine(cfg, rs).WithFirst(flagSet)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic")
		}
	}()
	en.WithFirst(flagSet)
}

func TestDbEngineRoot(t *testing.T) {
	nul := getNull()
	defer nul.Close()
	ctx := context.Background()
	cfg := Config{}
	rs := resource.NewMenuResource()
	rs.WithCodeGetter(codeGet)
	en := NewEngine(cfg, rs)
	cont, err := en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Fatalf("expected continue")
	}

	cont, err = en.Exec(ctx, []byte{0x30})
	if err == nil {
		t.Fatalf("expected loadfail")
	}

	_, err = en.WriteResult(ctx, nul) 
	if err != nil {
		t.Fatal(err)
	}

	cont, err = en.Exec(ctx, []byte{0x30})
	if err == nil {
		t.Fatalf("expected nocode")
	}
	err = en.Finish()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDbEnginePersist(t *testing.T) {
	nul := getNull()
	defer nul.Close()
	ctx := context.Background()
	cfg := Config{
		FlagCount: 1,
		SessionId: "bar",
	}
	store := memdb.NewMemDb()
	store.Connect(ctx, "")
	pe := persist.NewPersister(store)
	rs := resource.NewMenuResource()
	rs.WithCodeGetter(codeGet)
	rs.AddLocalFunc("foo", flagSet)
	en := NewEngine(cfg, rs)
	en = en.WithPersister(pe)
	cont, err := en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Fatalf("expected continue")
	}

	cont, err = en.Exec(ctx, []byte{0x30})
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.WriteResult(ctx, nul) 
	if err != nil {
		t.Fatal(err)
	}
	err = en.Finish()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDbEngineDebug(t *testing.T) {
	w := bytes.NewBuffer(nil)
	ctx := context.Background()
	cfg := Config{
		Root: "tinkywinky",
		FlagCount: 1,
	}
	rs := resource.NewMenuResource()
	rs = rs.WithCodeGetter(codeGet)
	rs.AddLocalFunc("foo", flagSet)
	dbg := NewSimpleDebug(w)
	en := NewEngine(cfg, rs).WithDebug(dbg)
	c, err := en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !c {
		t.Fatal("expected true")
	}
	if len(w.Bytes()) == 0 {
		t.Fatal("expected non-empty debug")
	}
}

func TestDbConfigString(t *testing.T) {
	cfg := Config{
		Root: "tinkywinky",
	}
	s := cfg.String()
	if !strings.Contains(s, "tinky") {
		t.Fatalf("expected contains 'tinky', got: '%s'", s)
	}
}
