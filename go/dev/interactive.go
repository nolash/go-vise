package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"strings"
	"os"

	"git.defalsify.org/festive/engine"
)

func main() {
	var dir string
	var root string
	flag.StringVar(&dir, "d", ".", "resource dir to read from")
	flag.StringVar(&root, "root", "root", "entry point symbol")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	en := engine.NewDefaultEngine(dir)
	err := en.Init(root, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot init: %v\n", err)
		os.Exit(1)
	}

	b := bytes.NewBuffer(nil)
	en.WriteResult(b)
	fmt.Println(b.String())

	running := true
	for running {
		reader := bufio.NewReader(os.Stdin)
		in, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read input: %v\n", err)
			os.Exit(1)
		}
		in = strings.TrimSpace(in)
		running, err = en.Exec([]byte(in), ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "execution terminated: %v\n", err)
			os.Exit(1)
		}
		b := bytes.NewBuffer(nil)
		en.WriteResult(b)
		fmt.Println(b.String())
	}

}
