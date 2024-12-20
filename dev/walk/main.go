package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.defalsify.org/vise.git/debug"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/logging"
	fsdb "git.defalsify.org/vise.git/db/fs"

)

var (
	logg = logging.NewVanilla()
)

func main() {
	var dir string
	var root string

	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	rsStore := fsdb.NewFsDb()
	err := rsStore.Connect(ctx, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resource db connect error: %v", err)
		os.Exit(1)
	}

	rs := resource.NewDbResource(rsStore)

	nm := debug.NewNodeMap(root)
	err = nm.Run(ctx, rs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "node tree process fail: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s", nm)
}
