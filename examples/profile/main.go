package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise/cache"
	"git.defalsify.org/vise/engine"
	"git.defalsify.org/vise/resource"
	"git.defalsify.org/vise/state"
)

const (
	USERFLAG_IDENTIFIED = iota + 8
	USERFLAG_HAVENAME 
	USERFLAG_HAVEEMAIL
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "profile")
	emptyResult = resource.Result{}
)

type profileResource struct {
	*resource.FsResource
	st *state.State
	haveEntered bool
}

func newProfileResource(st *state.State, rs *resource.FsResource) *profileResource {
	return &profileResource{
		rs,
		st,
		false,
	}
}

func(pr *profileResource) checkEntry() error {
	if pr.haveEntered {
		return nil
	}
	one, err := pr.st.GetFlag(USERFLAG_HAVENAME)
	if err != nil {
		return err
	}
	two, err := pr.st.GetFlag(USERFLAG_HAVEEMAIL)
	if err != nil {
		return err
	}
	if one && two {
		_, err = pr.st.SetFlag(USERFLAG_IDENTIFIED)
		if err != nil {
			return err
		}
		pr.haveEntered = true
	}
	return nil
}

func(pr profileResource) nameSave(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	log.Printf("writing name to file")
	fp := path.Join(scriptDir, "myname.txt")
	err := ioutil.WriteFile(fp, input, 0600)
	if err != nil {
		return emptyResult, err
	}
	changed, err := pr.st.SetFlag(USERFLAG_HAVENAME)
	if err != nil {
		return emptyResult, err
	}
	if changed {
		pr.checkEntry()
	}
	return emptyResult, err
}

func(pr profileResource) emailSave(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	log.Printf("writing email to file")
	fp := path.Join(scriptDir, "myemail.txt")
	err := ioutil.WriteFile(fp, input, 0600)
	if err != nil {
		return emptyResult, err
	}
	changed, err := pr.st.SetFlag(USERFLAG_HAVEEMAIL)
	if err != nil {
		return emptyResult, err
	}
	if changed {
		pr.checkEntry()
	}
	return resource.Result{}, err
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
	rsf := resource.NewFsResource(scriptDir)
	rs := newProfileResource(&st, &rsf)
	rs.AddLocalFunc("do_name_save", rs.nameSave)
	rs.AddLocalFunc("do_email_save", rs.emailSave)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
		OutputSize: uint32(size),
	}
	ctx := context.Background()
	en, err := engine.NewEngine(cfg, &st, rs, ca, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine create fail: %v\n", err)
		os.Exit(1)
	}

	err = engine.Loop(&en, os.Stdin, os.Stdout, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
