package engine

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"git.defalsify.org/vise/cache"
	"git.defalsify.org/vise/resource"
	"git.defalsify.org/vise/state"
	"git.defalsify.org/vise/testdata"
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
	wr := FsWrapper {
		&rs, 
		st,
	}
	wr.AddLocalFunc("one", wr.one)
	wr.AddLocalFunc("inky", wr.inky)
	wr.AddLocalFunc("pinky", wr.pinky)
	return wr
}

func(fs FsWrapper) one(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	return resource.Result{
		Content: "one",
	}, nil
}

func(fs FsWrapper) inky(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	return resource.Result{
		Content: "tinkywinky",
	}, nil
}

func(fs FsWrapper) pinky(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	r := fmt.Sprintf("xyzzy: %x", input)
	return resource.Result{
		Content: r,
	}, nil
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
	var err error
	generateTestData(t)
	ctx := context.TODO()
	st := state.NewState(17)
	rs := NewFsWrapper(dataDir, &st)
	ca := cache.NewCache().WithCacheSize(1024)
	
	en := NewEngine(Config{
		Root: "root",
	}, &st, &rs, ca, ctx)
//
	w := bytes.NewBuffer(nil)
	_, err = en.WriteResult(w, ctx)
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
	_, err = en.WriteResult(w, ctx)
	if err != nil {
		t.Fatal(err)
	}
	b = w.Bytes()
	expect := `this is in foo

it has more lines
0:to foo
1:go bar
2:see long`

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

	en := NewEngine(Config{}, &st, &rs, ca, ctx)
	err := en.Init("root", ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = en.Exec([]byte("_foo"), ctx)
	if err == nil {
		t.Fatalf("expected fail on invalid input")
	}
}
