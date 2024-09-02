// Example: Assemble and retrieve state flags using string identifiers specified in csv file.
package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	fsdb "git.defalsify.org/vise.git/db/fs"
)


var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "preprocessor")
)

type countResource struct {
	parser *asm.FlagParser
	count int
}

func newCountResource(fp string) (*countResource, error) {
	var err error
	pfp := path.Join(fp, "pp.csv")
	parser := asm.NewFlagParser()
	_, err = parser.Load(pfp)
	if err != nil {
		return nil, err
	}
	return &countResource{
		count: 0,
		parser: parser,
	}, nil
}

func(rsc* countResource) poke(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var r resource.Result

	ss := strings.Split(sym, "_")

	r.Content = "You will see this if this flag did not have a description"
	r.FlagReset = []uint32{8, 9, 10}
	v, err := rsc.parser.GetFlag(ss[1])
	if err != nil {
		v = 8 + uint32(rsc.count) + 1
		r.FlagSet = []uint32{8 + uint32(rsc.count) + 1}
	}
	r.FlagSet = []uint32{uint32(v)}
	s, err := rsc.parser.GetDescription(v)
	if err == nil {
		r.Content = s 
	}

	rsc.count++

	return r, nil
}

func main() {
	root := "root"
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	ctx := context.Background()
	st := state.NewState(5)
	st.UseDebug()
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rsf := resource.NewDbResource(store)
	rs, err := newCountResource(scriptDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aux handler fail: %v\n", err)
		os.Exit(1)
	}
	rsf.AddLocalFunc("flag_foo", rs.poke)
	rsf.AddLocalFunc("flag_bar", rs.poke)
	rsf.AddLocalFunc("flag_schmag", rs.poke)
	rsf.AddLocalFunc("flag_start", rs.poke)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
	}
	en := engine.NewEngine(cfg, rsf)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	
	_, err = en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init fail: %v\n", err)
		os.Exit(1)
	}

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
