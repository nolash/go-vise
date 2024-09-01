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
	fsdb "git.defalsify.org/vise.git/db/fs"
)
var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "longmenu")
)

func main() {
	var root string
	var size uint
	var sessionId string
	var persist bool
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.BoolVar(&persist, "persist", false, "use state persistence")
	flag.Parse()

	dir := scriptDir
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	dp := path.Join(scriptDir, ".state")
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, dp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect error: %v", err)
		os.Exit(1)
	}
	defer store.Close()
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
