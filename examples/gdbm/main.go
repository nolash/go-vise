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

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "gdbm")
	dbFile = path.Join(scriptDir, "vise.gdbm")
)

func do(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "bye",
	}, nil
}
	     
func main() {
	var err error
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	st := state.NewState(0)
	rs := resource.NewGdbmResource(dbFile)
	ca := cache.NewCache()
	if err != nil {
		panic(err)
	}
	cfg := engine.Config{
		Root: "root",
		Language: "nor",
	}
	ctx := context.Background()
	en := engine.NewEngine(ctx, cfg, &st, rs, ca)

	rs.AddLocalFunc("do", do)

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
