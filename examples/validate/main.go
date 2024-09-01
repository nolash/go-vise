// Example: Input checker.
package main

import (
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
	"git.defalsify.org/vise.git/db"
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "validate")
	emptyResult = resource.Result{}
)

const (
	USERFLAG_HAVESOMETHING = state.FLAG_USERSTART
)

type verifyResource struct {
	*resource.DbResource
	st *state.State
}

func(vr *verifyResource) verify(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var err error
	if string(input) == "something" {
		vr.st.SetFlag(USERFLAG_HAVESOMETHING)
	}
	return resource.Result{
		Content: "",
	}, err
}

func(vr *verifyResource) again(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	vr.st.ResetFlag(USERFLAG_HAVESOMETHING)
	return resource.Result{}, nil
}

func main() {
	var root string
	var size uint
	var sessionId string
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	st := state.NewState(1)
	store := db.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rsf := resource.NewDbResource(store)
	rs := verifyResource{rsf, &st}
	rs.AddLocalFunc("verifyinput", rs.verify)
	rs.AddLocalFunc("again", rs.again)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
		OutputSize: uint32(size),
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
