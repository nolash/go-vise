package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/db"
)

func main() {
	var store db.Db
	var dir string
	var root string
	var size uint
	var sessionId string
	var persist string
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.StringVar(&persist, "p", "", "state persistence directory")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	if persist != "" {
		store = &db.FsDb{}
		err := store.Connect(ctx, persist)
		if err != nil {
			fmt.Fprintf(os.Stderr, "db connect error: %v", err)
			os.Exit(1)
		}
	}
	en, err := engine.NewSizedEngine(dir, uint32(size), store, &sessionId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine create error: %v", err)
		os.Exit(1)
	}
	cont, err := en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init exited with error: %v\n", err)
		os.Exit(1)
	}
	if !cont {
		_, err = en.WriteResult(ctx, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dead init write error: %v\n", err)
			os.Exit(1)
		}
		os.Stdout.Write([]byte{0x0a})
		os.Exit(0)
	}
	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
