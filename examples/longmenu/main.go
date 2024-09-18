// Example: Pagination of long resource result content.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	fsdb "git.defalsify.org/vise.git/db/fs"
)
var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "longmenu")
)

func main() {
	var size uint
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.Parse()

	ctx := context.Background()
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect error: %v", err)
		os.Exit(1)
	}
	rs := resource.NewDbResource(store)
	defer rs.Close()
	cfg := engine.Config {
		OutputSize: uint32(size),
	}
	en := engine.NewEngine(cfg, rs)
	//cont, err := en.Init(ctx)
	cont, err := en.Exec(ctx, []byte{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init exited with error: %v\n", err)
		os.Exit(1)
	}
	if !cont {
		_, err = en.Flush(ctx, os.Stdout)
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
