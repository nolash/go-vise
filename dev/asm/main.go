package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"git.defalsify.org/vise/asm"
)

func main() {
	if (len(os.Args) < 2) {
		os.Exit(1)
	}
	fp := os.Args[1]
	v, err := ioutil.ReadFile(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v", err)
		os.Exit(1)
	}
	n, err := asm.Parse(string(v), os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v", err)
		os.Exit(1)
	}
	log.Printf("parsed total %v bytes", n)
}
