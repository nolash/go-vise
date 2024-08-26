package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)


var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "preprocessor")
	stringFlags = make(map[string]int)
)

type countResource struct {
	resource.Resource
	count int
}

func newCountResource() countResource {
	fs := resource.NewFsResource(scriptDir)
	return countResource{
		Resource: fs,
		count: 0,
	}
}

func(rsc* countResource) poke(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result

	ss := strings.Split(sym, "_")

	r.FlagReset = []uint32{8, 9, 10}
	v, ok := stringFlags[ss[1]]
	if ok {
		r.FlagSet = []uint32{uint32(v)}
	} else {
		r.FlagSet = []uint32{8 + uint32(rsc.count) + 1}
	}
	rsc.count++
	r.Content = "You will see this if no flag was set from code"
	engine.Logg.DebugCtxf(ctx, "countresource >>>>>> foo", "v", v, "ok", ok, "r", r)
	return r, nil
}

func main() {
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	stringFlags["foo"] = 8
	stringFlags["bar"] = 10

	st := state.NewState(5)
	st.UseDebug()
	rsf := resource.NewFsResource(scriptDir)
	rs := newCountResource()
	rsf.AddLocalFunc("flag_foo", rs.poke)
	rsf.AddLocalFunc("flag_bar", rs.poke)
	rsf.AddLocalFunc("flag_schmag", rs.poke)
	rsf.AddLocalFunc("flag_start", rs.poke)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
	}
	ctx := context.Background()
	en := engine.NewEngine(ctx, cfg, &st, rsf, ca)
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
