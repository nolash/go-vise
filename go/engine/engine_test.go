package engine

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

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

func (r FsWrapper) RenderTemplate(sym string, values map[string]string) (string, error) {
	return resource.DefaultRenderTemplate(r, sym, values)	
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
	st := state.NewState(17).WithCacheSize(1024)
	generateTestData(t)
	ctx := context.TODO()
	rs := NewFsWrapper(dataDir, &st)
	en := NewEngine(&st, &rs)
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
	if !bytes.Equal(b, []byte("hello world")) {
		t.Fatalf("expected result 'hello world', got %v", b)
	}

	input := []byte("1")
	err = en.Exec(input, ctx)
	if err != nil {
		t.Fatal(err)
	}
	r := st.Where()
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
