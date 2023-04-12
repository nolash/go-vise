package engine

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
)

func TestLoopBackForth(t *testing.T) {
	generateTestData(t)
	ctx := context.TODO()
	st := state.NewState(0)
	rs := resource.NewFsResource(dataDir)
	ca := cache.NewCache().WithCacheSize(1024)
	
	en := NewEngine(Config{}, &st, &rs, ca)
	err := en.Init("root", ctx)
	if err != nil {
		t.Fatal(err)
	}

	input := []string{
		"1",
		"0",
		"1",
		"0",
		}		
	inputStr := strings.Join(input, "\n")
	inputBuf := bytes.NewBuffer(append([]byte(inputStr), 0x0a))
	outputBuf := bytes.NewBuffer(nil)
	log.Printf("running with input: %s", inputBuf.Bytes())

	err = Loop(&en, "root", ctx, inputBuf, outputBuf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoopBrowse(t *testing.T) {
	generateTestData(t)
	ctx := context.TODO()
	st := state.NewState(0)
	rs := resource.NewFsResource(dataDir)
	ca := cache.NewCache().WithCacheSize(1024)

	cfg := Config{
		OutputSize: 68,
	}
	en := NewEngine(cfg, &st, &rs, ca)
	err := en.Init("root", ctx)
	if err != nil {
		t.Fatal(err)
	}

	input := []string{
		"1",
		"2",
		"00",
		"11",
		"00",
		}
	inputStr := strings.Join(input, "\n")
	inputBuf := bytes.NewBuffer(append([]byte(inputStr), 0x0a))
	outputBuf := bytes.NewBuffer(nil)
	log.Printf("running with input: %s", inputBuf.Bytes())

	err = Loop(&en, "root", ctx, inputBuf, outputBuf)
	if err != nil {
		t.Fatal(err)
	}

	location, idx := st.Where()
	if location != "long" {
		fmt.Errorf("expected location 'long', got %s", location)
	}
	if idx != 1 {
		fmt.Errorf("expected idx 1, got %v", idx)
	}
}