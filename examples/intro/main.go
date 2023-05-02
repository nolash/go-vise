package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG_HAVESOMETHING = iota + state.FLAG_USERSTART
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "intro")
)

type Counter struct {
	*resource.FsResource 
	c int64
	v string
}

func newCounter() Counter {
	fs := resource.NewFsResource(scriptDir)
	return Counter{fs, 0, ""}
}

func(c *Counter) count(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	s := "%v time"
	if c.c != 1 {
		s += "s"
	}
	r := resource.Result{
		Content: fmt.Sprintf(s, c.c),
	}
	c.c += 1 
	return  r, nil
}

func(c *Counter) something(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	c.v = string(input)	
	r := resource.Result{
		Content: c.v,
	}
	if len(input) > 0 {
		r.FlagSet = []uint32{USERFLAG_HAVESOMETHING}
	}
	return r, nil
}

func main() {
	var dir string
	var root string
	var size uint
	var sessionId string
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	st := state.NewState(3)
	rs := newCounter()
	rs.AddLocalFunc("count", rs.count)
	rs.AddLocalFunc("something", rs.something)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
		OutputSize: uint32(size),
	}
	ctx := context.Background()
	en := engine.NewEngine(ctx, cfg, &st, rs, ca)
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
