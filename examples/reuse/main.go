// Example: Reuse go functions for multiple LOAD symbols.
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
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG = iota + state.FLAG_USERSTART
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
	st := state.NewState(0)
	store := db.NewFsDb()
	store.Connect(ctx, scriptDir)
	rs, err := resource.NewDbResource(store, db.DATATYPE_TEMPLATE, db.DATATYPE_BIN, db.DATATYPE_MENU)
	rs.AddLocalFunc("do_foo", same)
	rs.AddLocalFunc("do_bar", same)
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
