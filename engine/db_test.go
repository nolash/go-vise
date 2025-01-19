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
	//cont, err := en.Init(ctx)
	cont, err := en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if cont {
		t.Fatalf("expected not continue")
	}
	err = en.Finish(ctx)
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
	//cont, err := en.Init(ctx)
	cont, err := en.Exec(ctx, []byte{})
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

	cont, err = en.Exec(ctx, []byte{0x30})
	if err == nil {
		t.Fatalf("expected nocode")
	}
	err = en.Finish(ctx)
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
	//cont, err := en.Init(ctx)
	cont, err := en.Exec(ctx, []byte{})
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

	_, err = en.Flush(ctx, nul) 
	if err != nil {
		t.Fatal(err)
	}
	err = en.Finish(ctx)
	if err != nil {
		t.Fatal(err)
	}

	en = NewEngine(cfg, rs)
	pe = persist.NewPersister(store)
	en = NewEngine(cfg, rs)
	en = en.WithPersister(pe)
	cont, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	stn := pe.GetState()
	if !stn.MatchFlag(state.FLAG_USERSTART, true) {
		t.Fatalf("expected userstart set, have state %v", stn)
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
	//c, err := en.Init(ctx)
	c, err := en.Exec(ctx, []byte{})
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

func TestDbEnsure(t *testing.T) {
	var err error
	var cfg Config
	ctx := context.Background()
	rs := resource.NewMenuResource()
	store := memdb.NewMemDb()
	store.Connect(ctx, "")
	pe := persist.NewPersister(store)
	en := NewEngine(cfg, rs).WithPersister(pe)
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if pe.GetState() == nil {
		t.Fatal("expected persister state")
	}
	if pe.GetMemory() == nil {
		t.Fatal("expected persister memory")
	}
}

func TestDbKeepPersisterContent(t *testing.T) {
	var err error
	var cfg Config
	ctx := context.Background()
	rs := resource.NewMenuResource()
	st := state.NewState(0)
	ca := cache.NewCache()
	store := memdb.NewMemDb()
	store.Connect(ctx, "")
	pe := persist.NewPersister(store)
	pe = pe.WithContent(st, ca)
	en := NewEngine(cfg, rs).WithPersister(pe)
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	pest := pe.GetState()
	if pest != st {
		t.Fatalf("expected persisted state %p same as engine %p", pest, st)
	}
	peca := pe.GetMemory()
	if peca != ca {
		t.Fatalf("expected persisted cache %p same as engine %p", peca, st)
	}
}

func TestDbKeepState(t *testing.T) {
	var err error
	var cfg Config
	ctx := context.Background()
	rs := resource.NewMenuResource()
	st := state.NewState(0)
	ca := cache.NewCache()
	store := memdb.NewMemDb()
	store.Connect(ctx, "")
	pe := persist.NewPersister(store)
	en := NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	en = en.WithPersister(pe)
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	pest := pe.GetState()
	if pest != st {
		t.Fatalf("expected persisted state %p same as engine preset %p", pest, st)
	}
	peca := pe.GetMemory()
	if peca != ca {
		t.Fatalf("expected persisted cache %p same as engine preset %p", peca, st)
	}
}

func TestDbFirst(t *testing.T) {
	var err error
	var cfg Config
	ctx := context.Background()
	rs := resource.NewMenuResource()
	st := state.NewState(1)
	store := memdb.NewMemDb()
	store.Connect(ctx, "")

	v := st.GetFlag(state.FLAG_USERSTART)
	if v {
		t.Fatal("expected flag unset")
	}
	en := NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithFirst(flagSet)
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	v = st.GetFlag(state.FLAG_USERSTART)
	if !v {
		t.Fatal("expected flag set")
	}
}

