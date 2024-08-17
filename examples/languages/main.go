package main

import (
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG_FLIP = iota + state.FLAG_USERSTART
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "languages")
)

type langController struct {
	State *state.State
}

func(l *langController) lang(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	lang := "nor"
	var rs resource.Result
	if l.State.MatchFlag(USERFLAG_FLIP, true) {
		lang = "eng"
		rs.FlagReset = append(rs.FlagReset, USERFLAG_FLIP)
	} else {
		rs.FlagSet = append(rs.FlagSet, USERFLAG_FLIP)
	}
	rs.Content = lang
	rs.FlagSet = append(rs.FlagSet, state.FLAG_LANG)
	return rs, nil
}

func msg(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result
	var code string
	engine.Logg.DebugCtxf(ctx, "in msg", "ctx", ctx, "val", code)
	if ctx.Value("Language") != nil {
		lang := ctx.Value("Language").(lang.Language)
		code = lang.Code
	}
	switch code {
	case "nor":
		r.Content = "Denne meldingen er fra en ekstern funksjon"
	default:
		r.Content = "This message is from an external function"
	}
	return r, nil
}

func empty(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "",
	}, nil
}

func main() {
	st := state.NewState(1)
	state.FlagDebugger.Register(USERFLAG_FLIP, "FLIP")
	rs := resource.NewFsResource(scriptDir)

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
	pr.State.UseDebug()

	aux := &langController{
		State: pr.State,
	}
	rs.AddLocalFunc("swaplang", aux.lang)
	rs.AddLocalFunc("msg", msg)
	rs.AddLocalFunc("empty", empty)

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
