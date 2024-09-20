// Example: Graceful termination that will be resumed from top on next execution.
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
	scriptDir = path.Join(baseDir, "examples", "quit")
)

func quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "quitter!",
	}, nil
}

func main() {
	ctx := context.Background()
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(store)
	cfg := engine.Config{
		Root: "root",
	}
	rs.AddLocalFunc("quitcontent", quit)

	en := engine.NewEngine(cfg, rs)
	err = engine.Loop(ctx, en, os.Stdin, os.Stdout, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
