package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.defalsify.org/vise/engine"
)

func main() {
	var dir string
	var root string
	var size uint
	var sessionId string
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	en := engine.NewSizedEngine(dir, uint32(size))
	err := en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init exited with error: %v\n", err)
		os.Exit(1)
	}
	err = engine.Loop(&en, os.Stdin, os.Stdout, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
