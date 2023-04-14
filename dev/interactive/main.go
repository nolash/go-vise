package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.defalsify.org/festive/engine"
)

func main() {
	var dir string
	var root string
	var size uint
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	en := engine.NewSizedEngine(dir, uint32(size))
	err := engine.Loop(&en, os.Stdin, os.Stdout, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v", err)
		os.Exit(1)
	}
}