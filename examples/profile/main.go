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

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "profile")
)

func nameSave(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	log.Printf("writing name to file")
	fp := path.Join(scriptDir, "myname.txt")
	err := ioutil.WriteFile(fp, input, 0600)
	return resource.Result{}, err
}

func emailSave(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	log.Printf("writing email to file")
	fp := path.Join(scriptDir, "myemail.txt")
	err := ioutil.WriteFile(fp, input, 0600)
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

	st := state.NewState(0)
	rs := resource.NewFsResource(scriptDir)
	rs.AddLocalFunc("do_name_save", nameSave)
	rs.AddLocalFunc("do_email_save", emailSave)
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: "root",
		SessionId: sessionId,
		OutputSize: uint32(size),
	}
	ctx := context.Background()
	en := engine.NewEngine(cfg, &st, &rs, ca, ctx)

	err := engine.Loop(&en, os.Stdin, os.Stdout, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v", err)
		os.Exit(1)
	}
}
