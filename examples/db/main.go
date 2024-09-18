// Example: Use db.Db provider for all local data.
//
// BUG: This will be stuck on _catch if quit before first write.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/logging"
)

var (
	logg = logging.NewVanilla()
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "db")
	store = fsdb.NewFsDb()
	pr = persist.NewPersister(store)
	data_selector = []byte("my_data")
)

func say(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result
	store.SetPrefix(db.DATATYPE_USERDATA)

	st := pr.GetState()
	if st.MatchFlag(state.FLAG_USERSTART, false) {
		r.FlagSet = []uint32{8}
		r.Content = "0"
		return r, nil
	}
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
	asm.Parse("MAP say\n", b)
	asm.Parse("MOUT quit 0\n", b)
	asm.Parse("HALT\n", b)
	asm.Parse("INCMP argh 0\n", b)
	asm.Parse("INCMP update *\n", b)
	store.SetPrefix(db.DATATYPE_BIN)
	err := store.Put(ctx, []byte("root"), b.Bytes())
	if err != nil {
		return err
	}

	b = bytes.NewBuffer(nil)
	asm.Parse("HALT\n", b)
	err = store.Put(ctx, []byte("argh"), b.Bytes())
	if err != nil {
		return err
	}

	b = bytes.NewBuffer(nil)
	asm.Parse("RELOAD say\n", b)
	asm.Parse("MOVE _\n", b)
	err = store.Put(ctx, []byte("update"), b.Bytes())
	if err != nil {
		return err
	}
	return nil
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
	ctx := context.Background()
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	dataDir := path.Join(scriptDir, ".store")
	err := store.Connect(ctx, dataDir)
	if err != nil {
		panic(err)
	}
	store.SetSession("xyzzy")
	store.SetLock(db.DATATYPE_TEMPLATE | db.DATATYPE_MENU | db.DATATYPE_BIN, false)
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
	store.SetLock(db.DATATYPE_TEMPLATE | db.DATATYPE_MENU | db.DATATYPE_BIN, true)

	tg := resource.NewDbResource(store)
	rs := resource.NewMenuResource()
	rs.WithTemplateGetter(tg.GetTemplate)
	rs.WithMenuGetter(tg.GetMenu)
	rs.WithCodeGetter(tg.GetCode)
	rs.AddLocalFunc("say", say)

	cfg := engine.Config{
		Root: "root",
		FlagCount: 1,
	}

	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)

	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init fail: %v\n", err)
		os.Exit(1)
	}
	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}

}
