// Example: Graceful termination that will be resumed from top on next execution.
package main

import (
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "quit")
)

func quit(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "quitter!",
	}, nil
}

func main() {
	st := state.NewState(0)
	st.UseDebug()
	ca := cache.NewCache()
	rs := resource.NewFsResource(scriptDir)
	cfg := engine.Config{
		Root: "root",
	}
	ctx := context.Background()
	en := engine.NewEngine(ctx, cfg, &st, rs, ca)

	rs.AddLocalFunc("quitcontent", quit)

	var err error
	_, err = en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init fail: %v\n", err)
		os.Exit(1)
	}

	err = engine.Loop(ctx, &en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
