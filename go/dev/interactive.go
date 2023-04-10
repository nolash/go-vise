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
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	en := engine.NewDefaultEngine(dir)
	err := engine.Loop(&en, root, ctx, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v", err)
		os.Exit(1)
	}
}
