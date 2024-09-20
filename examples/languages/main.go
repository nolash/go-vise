// Example: Set and apply language translation based on input, with and without Gettext.
package main

import (
	"context"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"
	gotext "gopkg.in/leonelquinteros/gotext.v1"

	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/logging"
)

const (
	USERFLAG_FLIP = iota + state.FLAG_USERSTART
)

var (
	logg = logging.NewVanilla()
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "languages")
	translationDir = path.Join(scriptDir, "locale")
)

func codeFromCtx(ctx context.Context) string {
	var code string
	logg.DebugCtxf(ctx, "in msg", "ctx", ctx, "val", code)
	if ctx.Value("Language") != nil {
		lang := ctx.Value("Language").(lang.Language)
		code = lang.Code
	}
	return code
}

type langController struct {
	translations map[string]gotext.Locale
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
	switch codeFromCtx(ctx)	{
	case "nor":
		r.Content = "Denne meldingen er fra en ekstern funksjon"
	default:
		r.Content = "This message is from an external function"
	}
	return r, nil
}

func(l *langController) moMsg(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result
	code := codeFromCtx(ctx)
	o := gotext.NewLocale(translationDir, code)
	o.AddDomain("default")	
	r.Content = o.Get("This message is translated using gettext")
	logg.DebugCtxf(ctx, "lang", "code", code, "translateor", o)
	return r, nil
}

func empty(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "",
	}, nil
}

func main() {
	ctx := context.Background()
	rsStore := fsdb.NewFsDb()
	err := rsStore.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(rsStore)

	cfg := engine.Config{
		Root: "root",
		SessionId: "default",
	}

	dp := path.Join(scriptDir, ".state")
	store := fsdb.NewFsDb()
	err = store.Connect(ctx, dp)
	if err != nil {
		logg.ErrorCtxf(ctx, "db connect fail", "err", err)
		os.Exit(1)
	}
	pr := persist.NewPersister(store)
	st := state.NewState(1)
	en := engine.NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithPersister(pr)

	aux := &langController{
		State: st,
	}
	rs.AddLocalFunc("swaplang", aux.lang)
	rs.AddLocalFunc("msg", msg)
	rs.AddLocalFunc("momsg", aux.moMsg)
	rs.AddLocalFunc("empty", empty)

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
