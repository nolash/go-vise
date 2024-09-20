// Example: Reuse go functions for multiple LOAD symbols.
package main

import (
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "reuse")
	emptyResult = resource.Result{}
)

func same(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "You came through the symbol " + sym,
	}, nil
}

func main() {
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	ctx := context.Background()
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(store)
	
	rs.AddLocalFunc("do_foo", same)
	rs.AddLocalFunc("do_bar", same)
	cfg := engine.Config{
		Root: "root",
	}
	en := engine.NewEngine(cfg, rs)
	err = engine.Loop(ctx, en, os.Stdin, os.Stdout, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
