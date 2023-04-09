package engine

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/testdata"
)

var (
	dataGenerated bool = false
	dataDir string = testdata.DataDir
)

type FsWrapper struct {
	*resource.FsResource
	st *state.State
}

func NewFsWrapper(path string, st *state.State) FsWrapper {
	rs := resource.NewFsResource(path)
	return FsWrapper {
		&rs, 
		st,
	}
}

func(fs FsWrapper) one(ctx context.Context) (string, error) {
	return "one", nil
}

func(fs FsWrapper) inky(ctx context.Context) (string, error) {
	return "tinkywinky", nil
}

func(fs FsWrapper) FuncFor(sym string) (resource.EntryFunc, error) {
	switch sym {
	case "one":
		return fs.one, nil
	case "inky":
		return fs.inky, nil
	}
	return nil, fmt.Errorf("function for %v not found", sym)
}

func(fs FsWrapper) GetCode(sym string) ([]byte, error) {
	sym += ".bin"
	fp := path.Join(fs.Path, sym)
	r, err := ioutil.ReadFile(fp)
	return r, err
}

func generateTestData(t *testing.T) {
	if dataGenerated {
		return
	}
	var err error
	dataDir, err = testdata.Generate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestEngineInit(t *testing.T) {
	generateTestData(t)
	ctx := context.TODO()
	st := state.NewState(17)
	rs := NewFsWrapper(dataDir, &st)
	ca := cache.NewCache().WithCacheSize(1024)
	
	en := NewEngine(&st, &rs, ca)
	err := en.Init("root", ctx)
	if err != nil {
		t.Fatal(err)
	}

	w := bytes.NewBuffer(nil)
	err = en.WriteResult(w)
	if err != nil {
		t.Fatal(err)
	}
	b := w.Bytes()
	expect_str := `hello world
1:do the foo
2:go to the bar`

	if !bytes.Equal(b, []byte(expect_str)) {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect_str, b)
	}

	input := []byte("1")
	_, err = en.Exec(input, ctx)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := st.Where()
	if r != "foo" {
		t.Fatalf("expected where-string 'foo', got %s", r)
	}
	w = bytes.NewBuffer(nil)
	err = en.WriteResult(w)
	if err != nil {
		t.Fatal(err)
	}
	b = w.Bytes()
	expect := `this is in foo

it has more lines
0:to foo
1:go bar`
	if !bytes.Equal(b, []byte(expect)) {
		t.Fatalf("expected\n\t%s\ngot:\n\t%s\n", expect, b)
	}
}

func TestEngineExecInvalidInput(t *testing.T) {
	generateTestData(t)
	ctx := context.TODO()
	st := state.NewState(17)
	rs := NewFsWrapper(dataDir, &st)
	ca := cache.NewCache().WithCacheSize(1024)

	en := NewEngine(&st, &rs, ca)
	err := en.Init("root", ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = en.Exec([]byte("_foo"), ctx)
	if err == nil {
		t.Fatalf("expected fail on invalid input")
	}
}
