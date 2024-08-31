// Example: Toggling states with external functions.
package main

import (
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USER_FOO = iota + state.FLAG_USERSTART
	USER_BAR
	USER_BAZ
)

var (
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

	engine.Logg.DebugCtxf(ctx, "in do", "sym", sym)

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

	st := state.NewState(3)
	st.UseDebug()
	rs := resource.NewFsResource(scriptDir)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
	}
	ctx := context.Background()
	en := engine.NewEngine(ctx, cfg, &st, rs, ca)
	en.SetDebugger(engine.NewSimpleDebug(nil))

	aux := &flagResource{st: &st}
	rs.AddLocalFunc("do_foo", aux.do)
	rs.AddLocalFunc("do_bar", aux.do)
	rs.AddLocalFunc("do_baz", aux.do)
	rs.AddLocalFunc("states", aux.get)

	state.FlagDebugger.Register(USER_FOO, "FOO")
	state.FlagDebugger.Register(USER_BAR, "BAR")
	state.FlagDebugger.Register(USER_BAZ, "BAZ")
	var err error
	_, err = en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init fail: %v\n", err)
		os.Exit(1)
	}
	err = engine.Loop(ctx, &en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
