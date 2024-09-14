// Example: Toggling states with external functions, with engine debugger.
package main

import (
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/logging"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

const (
	USER_FOO = iota + state.FLAG_USERSTART
	USER_BAR
	USER_BAZ
)

var (
	logg = logging.NewVanilla()
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "state")
)

type flagResource struct {
	st *state.State
}

func(f *flagResource) get(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: state.FlagDebugger.AsString(f.st.Flags, 3),
	}, nil		
}


func(f *flagResource) do(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result

	logg.DebugCtxf(ctx, "in do", "sym", sym)

	switch(sym) {
	case "do_foo":
		if f.st.MatchFlag(USER_FOO, false) {
			r.FlagSet = append(r.FlagSet, USER_FOO)
		} else {
			r.FlagReset = append(r.FlagReset, USER_FOO)
		}
	case "do_bar":
		if f.st.MatchFlag(USER_BAR, false) {
			r.FlagSet = append(r.FlagSet, USER_BAR)
		} else {
			r.FlagReset = append(r.FlagReset, USER_BAR)
		}
	case "do_baz":
		if f.st.MatchFlag(USER_BAZ, false) {
			r.FlagSet = append(r.FlagSet, USER_BAZ)
		} else {
			r.FlagReset = append(r.FlagReset, USER_BAZ)
		}
	}
	return r, nil
}
	     
func main() {
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	ctx := context.Background()
	st := state.NewState(3)
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(store)
	cfg := engine.Config{
		Root: "root",
	}

	aux := &flagResource{st: st}
	rs.AddLocalFunc("do_foo", aux.do)
	rs.AddLocalFunc("do_bar", aux.do)
	rs.AddLocalFunc("do_baz", aux.do)
	rs.AddLocalFunc("states", aux.get)

	state.FlagDebugger.Register(USER_FOO, "FOO")
	state.FlagDebugger.Register(USER_BAR, "BAR")
	state.FlagDebugger.Register(USER_BAZ, "BAZ")

	en := engine.NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithDebug(nil)

	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init fail: %v\n", err)
		os.Exit(1)
	}
	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
