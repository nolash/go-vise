package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

func main() {
	var dir string
	var root string
	var size uint
	var sessionId string
	var persistDir string
	var initial string
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.StringVar(&persistDir, "p", "", "state persistence directory")
	flag.StringVar(&initial, "initial", "", "initial input to pass to engine initialization")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	cfg := engine.Config{
		OutputSize: uint32(size),
		SessionId: sessionId,
	}

	rsStore := fsdb.NewFsDb()
	err := rsStore.Connect(ctx, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resource db connect error: %v", err)
		os.Exit(1)
	}

	rs := resource.NewDbResource(rsStore)
	rs = rs.With(db.DATATYPE_STATICLOAD)
	en := engine.NewEngine(cfg, rs)
	if persistDir != "" {
		store := fsdb.NewFsDb()
		err = store.Connect(ctx, persistDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "persist db connect error: %v", err)
			os.Exit(1)
		}
		pe := persist.NewPersister(store)
		en = en.WithPersister(pe)
	}

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout, []byte(initial))
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
