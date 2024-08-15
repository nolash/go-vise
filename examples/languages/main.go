package main

import (
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG_HAVESOMETHING = iota + state.FLAG_USERSTART
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "languages")
)

func lang(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "nor",
		FlagSet: []uint32{state.FLAG_LANG},
	}, nil
}

func main() {
	st := state.NewState(0)
	rs := resource.NewFsResource(scriptDir)
	rs.AddLocalFunc("swaplang", lang)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: "default",
	}
	ctx := context.Background()

	dp := path.Join(scriptDir, ".state")
	err := os.MkdirAll(dp, 0700)
	if err != nil {
		engine.Logg.ErrorCtxf(ctx, "cannot create state dir", "err", err)
		os.Exit(1)
	}
	pr := persist.NewFsPersister(dp)
	en, err := engine.NewPersistedEngine(ctx, cfg, pr, rs)
	if err != nil {
		engine.Logg.Infof("persisted engine create error. trying again with persisting empty state first...")
		pr = pr.WithContent(&st, ca)
		err = pr.Save(cfg.SessionId)
		if err != nil {
			engine.Logg.ErrorCtxf(ctx, "fail state save: %v", err)
			os.Exit(1)
		}
		en, err = engine.NewPersistedEngine(ctx, cfg, pr, rs)
	}

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
