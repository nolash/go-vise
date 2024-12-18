package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.defalsify.org/vise.git/vm"
	"git.defalsify.org/vise.git/resource"
	fsdb "git.defalsify.org/vise.git/db/fs"

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
	//rs = rs.With(db.DATATYPE_STATICLOAD)

	ph := vm.NewParseHandler().WithDefaultHandlers().WithWriter(os.Stdout)

	b, err := rs.GetCode(ctx, root)
	if err != nil {
		panic(err)
	}

	n, err := ph.ParseAll(b)
	if err != nil {
		panic(err)
	}
	_ = n
}
