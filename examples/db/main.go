package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/db"
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "db")
	ctx = context.Background()
	store = db.NewMemDb(ctx)
	data_selector = []byte("my_data")
)

func say(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result
	store.SetPrefix(db.DATATYPE_USERSTART)

	if len(input) > 0 {
		err := store.Put(ctx, data_selector, input)
		if err != nil {
			return r, err
		}
	}

	v, err := store.Get(ctx, data_selector)
	if err != nil {
		return r, err
	}

	r.Content = string(v)
	return r, nil
}

func genCode(ctx context.Context, store db.Db) error {
	b := bytes.NewBuffer(nil)
	asm.Parse("LOAD say 0\n", b)
	asm.Parse("RELOAD say\n", b)
	asm.Parse("MAP say\n", b)
	asm.Parse("MOUT quit 0\n", b)
	asm.Parse("HALT\n", b)
	asm.Parse("INCMP argh 0\n", b)
	asm.Parse("INCMP ^ *\n", b)
	store.SetPrefix(db.DATATYPE_BIN)
	err := store.Put(ctx, []byte("root"), b.Bytes())
	if err != nil {
		return err
	}

	b = bytes.NewBuffer(nil)
	asm.Parse("HALT\n", b)
	return store.Put(ctx, []byte("argh"), b.Bytes())
}

func genMenu(ctx context.Context, store db.Db) error {
	store.SetPrefix(db.DATATYPE_MENU)
	return store.Put(ctx, []byte("quit"), []byte("give up"))
}

func genTemplate(ctx context.Context, store db.Db) error {
	store.SetPrefix(db.DATATYPE_TEMPLATE)
	return store.Put(ctx, []byte("root"), []byte("current data is {{.say}}"))
}

func main() {
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	st := state.NewState(1)

	store.SetSession("xyzzy")
	store.SetPrefix(db.DATATYPE_USERSTART)
	err := store.Put(ctx, data_selector, []byte("0"))
	if err != nil {
		panic(err)
	}

	err = genCode(ctx, store)
	if err != nil {
		panic(err)
	}

	err = genMenu(ctx, store)
	if err != nil {
		panic(err)
	}

	err = genTemplate(ctx, store)
	if err != nil {
		panic(err)
	}

	tg, err := resource.NewDbFuncGetter(store, db.DATATYPE_TEMPLATE, db.DATATYPE_MENU, db.DATATYPE_BIN)
	if err != nil {
		panic(err)
	}
	rs := resource.NewMenuResource()
	rs.WithTemplateGetter(tg.GetTemplate)
	rs.WithMenuGetter(tg.GetMenu)
	rs.WithCodeGetter(tg.GetCode)
	rs.WithCodeGetter(tg.GetCode)
	rs.AddLocalFunc("say", say)

	ca := cache.NewCache()
	if err != nil {
		panic(err)
	}
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
