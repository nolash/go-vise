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
	"git.defalsify.org/vise.git/db"
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
	ctx := context.Background()
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	st := state.NewState(0)
	store := &db.GdbmDb{}
	err := store.Connect(ctx, dbFile)
	if err != nil {
		panic(err)
	}

	tg, err := resource.NewDbFuncGetter(store, db.DATATYPE_TEMPLATE, db.DATATYPE_BIN)
	if err != nil {
		panic(err)
	}
	rs := resource.NewMenuResource()
	rs = rs.WithTemplateGetter(tg.GetTemplate)
	rs = rs.WithCodeGetter(tg.GetCode)

	rsf := resource.NewFsResource(scriptDir)
	rsf.AddLocalFunc("do", do)
	rs = rs.WithMenuGetter(rsf.GetMenu)
	rs = rs.WithEntryFuncGetter(rsf.FuncFor)

	ca := cache.NewCache()
	if err != nil {
		panic(err)
	}
	cfg := engine.Config{
		Root: "root",
		Language: "nor",
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
