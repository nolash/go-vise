// Example: Sub-machine using the first function feature in engine
package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"crypto/rand"
	"crypto/sha256"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/logging"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

const (
	USER_CHALLENGED = iota + state.FLAG_USERSTART
	USER_FAILED
	USER_SUCCEEDED
)

var (
	logg = logging.NewVanilla()
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "first")
	storeDir = path.Join(scriptDir, ".state")
	authStoreDir = path.Join(scriptDir, ".auth")
)

type firstAuthResource struct {
	st *state.State
	store db.Db
}

func(f *firstAuthResource) challenge(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	var err error
	var r resource.Result

	if len(input) == 0 {
		return r, nil
	}
	if input[0] != 0x25 { // %
		return r, nil
	}

	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return r, errors.New("missing session")
	}

	if !f.st.GetFlag(USER_CHALLENGED) {
		succeed := f.st.GetFlag(USER_SUCCEEDED)
		failed := f.st.GetFlag(USER_FAILED)
		if succeed {
			return r, nil	
		} else {
			if failed {
				r.FlagReset = append(r.FlagReset, USER_FAILED)
			}
			b := make([]byte, 32)
			_, err = rand.Read(b)
			if err != nil {
				return r, err
			}
			f.store.SetPrefix(db.DATATYPE_USERDATA)
			f.store.SetSession(sessionId)
			err := f.store.Put(ctx, []byte("challenge"), b)
			if err != nil {
				return r, err
			}
			r.FlagSet = append(r.FlagSet, USER_CHALLENGED)
			r.FlagSet = append(r.FlagSet, state.FLAG_TERMINATE)
			r.Content = hex.EncodeToString(b)
			return r, nil
		}
	} else {
		logg.DebugCtxf(ctx, "have challenge response", "input", input)
		f.store.SetPrefix(db.DATATYPE_USERDATA)
		f.store.SetSession(sessionId)
		v, err := f.store.Get(ctx, []byte("challenge"))
		if err != nil {
			return r, err
		}
		h := sha256.New()
		h.Write([]byte(sessionId))
		h.Write(v)
		b := h.Sum(nil)
		x := hex.EncodeToString(b)
		if x != string(input[1:]) {
			r.FlagSet = append(r.FlagSet, USER_FAILED)
			r.FlagSet = append(r.FlagSet, state.FLAG_TERMINATE)
			r.FlagReset = append(r.FlagReset, USER_CHALLENGED)
		} else {
			r.FlagSet = append(r.FlagSet, USER_SUCCEEDED)
		}
	}
	return r, nil
}
	     
func main() {
	var cont bool
	var sessionId string
	var input string
	root := "root"
	flag.StringVar(&sessionId, "session-id", "default", "session id")
	flag.Parse()
	input = flag.Arg(0)
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, scriptDir)

	ctx := context.Background()
	st := state.NewState(3)
	st.UseDebug()
	state.FlagDebugger.Register(USER_CHALLENGED, "AUTHCHALLENGED")
	state.FlagDebugger.Register(USER_FAILED, "AUTHFAILED")
	state.FlagDebugger.Register(USER_SUCCEEDED, "AUTHSUCCEEDED")
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	stateStore := fsdb.NewFsDb()
	err = stateStore.Connect(ctx, storeDir)
	if err != nil {
		panic(err)
	}
	authStore := fsdb.NewFsDb()
	err = authStore.Connect(ctx, authStoreDir)
	if err != nil {
		panic(err)
	}

	rs := resource.NewDbResource(store)
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
	}

	aux := &firstAuthResource{
		st: st,
		store: authStore,
	}

	pe := persist.NewPersister(stateStore)
	en := engine.NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithFirst(aux.challenge)
	en = en.WithPersister(pe)
	err = en.AddValidInput("^%.*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine inputvalid add fail: %v\n", err)
		os.Exit(1)
	}

	cont, err = en.Exec(ctx, []byte(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init pre fail: %v\n", err)
		os.Exit(1)
	}
	if cont {
		_, err = en.Exec(ctx, []byte(input))
		if err != nil {
			fmt.Fprintf(os.Stderr, "engine init after fail: %v\n", err)
			os.Exit(1)
		}
	}

	_, err = en.Flush(ctx, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine finish fail: %v\n", err)
		os.Exit(1)
	}

	err = en.Finish(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine finish fail: %v\n", err)
		os.Exit(1)
	}
}
