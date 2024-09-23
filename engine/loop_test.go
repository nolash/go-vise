package engine

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
)

func TestLoopTop(t *testing.T) {
	var err error
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(0)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache().WithCacheSize(1024)

	cfg := Config{
		Root: "root",
	}
	en := NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)

	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	input := []string{
		"2",
		"j",
		"1",
		}		
	inputStr := strings.Join(input, "\n")
	inputBuf := bytes.NewBuffer(append([]byte(inputStr), 0x0a))
	outputBuf := bytes.NewBuffer(nil)
	log.Printf("running with input: %s", inputBuf.Bytes())

	err = Loop(ctx, en, inputBuf, outputBuf, nil)
	if err != nil {
		t.Fatal(err)
	}
	location, _ := st.Where()
	if location != "foo" {
		fmt.Errorf("expected location 'foo', got %s", location)
	}
}

func TestLoopBackForth(t *testing.T) {
	var err error
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(0)
	rs := newTestWrapper(dataDir, st)
	
	cfg := Config{
		Root: "root",
		CacheSize: 1024,
	}
	en := NewEngine(cfg, rs)
	en = en.WithState(st)

	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
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

	err = Loop(ctx, en, inputBuf, outputBuf, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoopBrowse(t *testing.T) {
	var err error
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(0)
	rs := newTestWrapper(dataDir, st)

	cfg := Config{
		OutputSize: 68,
		Root: "root",
		CacheSize: 1024,
	}
	en := NewEngine(cfg, rs)
	en = en.WithState(st)

	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	input := []string{
		"2",
		"00",
		"11",
		"00",
		}
	inputStr := strings.Join(input, "\n")
	inputBuf := bytes.NewBuffer(append([]byte(inputStr), 0x0a))
	outputBuf := bytes.NewBuffer(nil)
	log.Printf("running with input: %s", inputBuf.Bytes())

	err = Loop(ctx, en, inputBuf, outputBuf, []byte("1"))
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
