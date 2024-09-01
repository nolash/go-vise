// Example: States and branching to check a PIN for access.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/db"
)

const (
	USERFLAG_VALIDPIN = iota + state.FLAG_USERSTART
	USERFLAG_QUERYPIN
)

var (
	logg = logging.NewVanilla()
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "pincheck")
	pin = []byte("1234")
)

type pinResource struct{
	resource.Resource
	st *state.State
}

func newPinResource(resource resource.Resource, state *state.State) *pinResource {
	return &pinResource{
		resource,
		state,
	}
}

func(rs *pinResource) pinCheck(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result

	if rs.st.MatchFlag(USERFLAG_QUERYPIN, false) {
		r.Content = "Please enter PIN"
		r.FlagReset = []uint32{USERFLAG_VALIDPIN}
		r.FlagSet = []uint32{USERFLAG_QUERYPIN}
		return r, nil
	}
	if bytes.Equal(input, pin) {
		r.FlagSet = []uint32{USERFLAG_VALIDPIN}
		logg.DebugCtxf(ctx, "pin match", "state", rs.st, "rs", r.FlagSet, "rr", r.FlagReset)
	} else {
		r.Content = "Wrong PIN please try again"
	}
	return r, nil
}

func(rs *pinResource) pinClear(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result
	r.FlagReset = []uint32{USERFLAG_VALIDPIN, USERFLAG_QUERYPIN}
	return r, nil
}

func main() {
	root := "root"
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	ctx := context.Background()
	st := state.NewState(3)
	st.UseDebug()
	state.FlagDebugger.Register(USERFLAG_VALIDPIN, "VALIDPIN")
	state.FlagDebugger.Register(USERFLAG_QUERYPIN, "QUERYPIN")
	store := db.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rsf := resource.NewDbResource(store)
	rs := newPinResource(rsf, &st)
	rsf.AddLocalFunc("pincheck", rs.pinCheck)
	rsf.AddLocalFunc("pinclear", rs.pinClear)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
	}
	en := engine.NewEngine(ctx, cfg, &st, rs, ca)
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
