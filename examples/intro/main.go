// Example: Basic flags and input processing, and symbol execution.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

const (
	USERFLAG_HAVESOMETHING = iota + state.FLAG_USERSTART
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "intro")
)

type introResource struct {
	*resource.DbResource
	c int64
	v []string
}

func newintroResource(ctx context.Context) introResource {
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir) 
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(store)
	return introResource{rs, 0, []string{}}
}

// increment counter.
// return a string representing the current value of the counter.
func(c *introResource) count(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

// if input is suppled, append it to the stored string vector and set the HAVESOMETHING flag.
// return the stored string vector value, one string per line.
func(c *introResource) something(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	c.v = append(c.v, string(input))
	r := resource.Result{
		Content: strings.Join(c.v, "\n"),
	}
	if len(input) > 0 {
		r.FlagSet = []uint32{USERFLAG_HAVESOMETHING}
	}
	return r, nil
}

func main() {
	var err error
	var dir string
	var root string
	var size uint
	var sessionId string
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)
	
	ctx := context.Background()
	st := state.NewState(3)
	rs := newintroResource(ctx)
	rs.AddLocalFunc("count", rs.count)
	rs.AddLocalFunc("something", rs.something)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
		OutputSize: uint32(size),
	}
	en := engine.NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)

	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
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
