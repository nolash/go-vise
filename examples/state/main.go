package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
)

const (
	USERFLAG_ONE = iota + state.FLAG_USERSTART
	USERFLAG_TWO
	USERFLAG_THREE
	USERFLAG_DONE
)

type fsData struct {
	path string
	state *state.State
}

func (fsd *fsData) poke(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return res, err
	}
	f.Write([]byte("*"))
	f.Close()
	f, err = os.Open(fp)
	if err != nil {
		return res, err
	}
	r, err := ioutil.ReadAll(f)
	if err != nil {
		return res, err
	}
	st := fsd.state
	for i := 8; i < 12; i++ {
		v := uint32(i)
		if st.MatchFlag(v, true) {
			res.FlagReset = append(res.FlagReset, v)
			res.FlagSet = append(res.FlagSet, v + 1)
			break
		}
	}
	if len(res.FlagSet) == 0 {
		res.FlagSet = append(res.FlagSet, 8)
	}
	res.FlagSet = append(res.FlagSet, state.FLAG_DIRTY)
	res.Content = string(r)
	return res, nil
}

func main() {
	var dir string
	var root string
	var size uint
	var sessionId string
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.UintVar(&size, "s", 0, "max size of output")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	st := state.NewState(4)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
	}

	dp := path.Join(dir, ".state")
	
	err := os.MkdirAll(dp, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "state dir create exited with error: %v\n", err)
		os.Exit(1)
	}
	pr := persist.NewFsPersister(dp)
	en, err := engine.NewPersistedEngine(ctx, cfg, pr, rs)
	if err != nil {
		pr = pr.WithContent(&st, ca)
		err = pr.Save(cfg.SessionId)
		en, err = engine.NewPersistedEngine(ctx, cfg, pr, rs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "engine create exited with error: %v\n", err)
			os.Exit(1)
		}
	}

	fp := path.Join(dp, sessionId)
	aux := &fsData{
		path: fp,
		state: &st,
	}
	rs.AddLocalFunc("count", aux.poke)

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
		err = en.Finish()
		if err != nil {
			fmt.Fprintf(os.Stderr, "engine finish error: %v\n", err)
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
