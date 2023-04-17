package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise/cache"
	"git.defalsify.org/vise/engine"
	"git.defalsify.org/vise/resource"
	"git.defalsify.org/vise/state"
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
	*resource.FsResource
	st *state.State
}

func(vr *verifyResource) verify(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	var err error
	if string(input) == "something" {
		_, err = vr.st.SetFlag(USERFLAG_HAVESOMETHING)
	}
	return resource.Result{
		Content: "",
	}, err
}

func(vr *verifyResource) again(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	var err error
	_, err = vr.st.ResetFlag(USERFLAG_HAVESOMETHING)
	return resource.Result{}, err
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

	st := state.NewState(1)
	rsf := resource.NewFsResource(scriptDir)
	rs := verifyResource{&rsf, &st}
	rs.AddLocalFunc("verifyinput", rs.verify)
	rs.AddLocalFunc("again", rs.again)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
		OutputSize: uint32(size),
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	en := engine.NewEngine(cfg, &st, rs, ca, ctx)
	var err error
	_, err = en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init fail: %v\n", err)
		os.Exit(1)
	}
	err = engine.Loop(&en, os.Stdin, os.Stdout, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
