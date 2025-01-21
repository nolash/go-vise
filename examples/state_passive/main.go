// Example: Asynchronous state persistence.
//
// TODO: this example is broken after change to loop
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG_ONE = iota + state.FLAG_USERSTART
	USERFLAG_TWO
	USERFLAG_THREE
	USERFLAG_DONE
)

var (
	logg = logging.NewVanilla()
)

type fsData struct {
	path      string
	persister *persist.Persister
}

func (fsd *fsData) peek(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	res := resource.Result{}
	fp := fsd.path + "_data"
	f, err := os.Open(fp)
	if err != nil {
		return res, err
	}
	r, err := ioutil.ReadAll(f)
	if err != nil {
		return res, err
	}
	res.Content = string(r)
	return res, nil
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

	st := fsd.persister.GetState()
	for i := 8; i < 12; i++ {
		v := uint32(i)
		logg.DebugCtxf(ctx, "checking flag", "flag", v)
		if st.MatchFlag(v, true) {
			logg.DebugCtxf(ctx, "match on flag", "flag", v)
			res.FlagReset = append(res.FlagReset, v)
			res.FlagSet = append(res.FlagSet, v+1)
			break
		}
	}
	if len(res.FlagSet) == 0 {
		res.FlagSet = append(res.FlagSet, 8)
	}
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
	rsStore := fsdb.NewFsDb()
	err := rsStore.Connect(ctx, dir)
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(rsStore)
	cfg := engine.Config{
		Root:      "root",
		SessionId: sessionId,
		FlagCount: 4,
	}

	dp := path.Join(dir, ".state")
	store := fsdb.NewFsDb()
	err = store.Connect(ctx, dp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect fail: %s", err)
		os.Exit(1)
	}
	pr := persist.NewPersister(store)

	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)

	fp := path.Join(dp, sessionId)
	aux := &fsData{
		path:      fp,
		persister: pr,
	}
	rs.AddLocalFunc("poke", aux.poke)
	rs.AddLocalFunc("peek", aux.peek)

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
