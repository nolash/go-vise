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

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

const (
	USERFLAG_IDENTIFIED = iota + state.FLAG_USERSTART
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

func newProfileResource(st *state.State, rs *resource.FsResource) resource.Resource {
	return &profileResource{
		rs,
		st,
		false,
	}
}

func(pr *profileResource) checkEntry() error {
	log.Printf("%v %v", USERFLAG_IDENTIFIED, USERFLAG_HAVENAME)
	if pr.haveEntered {
		return nil
	}
	one := pr.st.GetFlag(USERFLAG_HAVENAME)
	two := pr.st.GetFlag(USERFLAG_HAVEEMAIL)
	if one && two {
		pr.st.SetFlag(USERFLAG_IDENTIFIED)
		pr.haveEntered = true
	}
	return nil
}

func(pr profileResource) nameSave(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	log.Printf("writing name to file")
	fp := path.Join(scriptDir, "myname.txt")
	err := ioutil.WriteFile(fp, input, 0600)
	if err != nil {
		return emptyResult, err
	}
	changed := pr.st.SetFlag(USERFLAG_HAVENAME)
	if changed {
		pr.checkEntry()
	}
	return emptyResult, err
}

func(pr profileResource) emailSave(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	log.Printf("writing email to file")
	fp := path.Join(scriptDir, "myemail.txt")
	err := ioutil.WriteFile(fp, input, 0600)
	if err != nil {
		return emptyResult, err
	}
	changed := pr.st.SetFlag(USERFLAG_HAVEEMAIL)
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
	rs, ok := newProfileResource(&st, rsf).(*profileResource)
	if !ok {
		os.Exit(1)
	}
	rs.AddLocalFunc("do_name_save", rs.nameSave)
	rs.AddLocalFunc("do_email_save", rs.emailSave)
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
