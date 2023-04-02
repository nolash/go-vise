package engine

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"text/template"
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

func NewFsWrapper(path string, st *state.State, ctx context.Context) FsWrapper {
	rs := resource.NewFsResource(path, ctx)
	return FsWrapper {
		&rs, 
		st,
	}
}

func (r FsWrapper) RenderTemplate(sym string, values map[string]string) (string, error) {
	v, err := r.GetTemplate(sym)
	if err != nil {
		return "", err
	}
	tp, err := template.New("tester").Option("missingkey=error").Parse(v)
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer([]byte{})
	err = tp.Execute(b, values)
	if err != nil {
		return "", err
	}
	return b.String(), err
}

func(fs FsWrapper) one(ctx context.Context) (string, error) {
	return "one", nil
}

func(fs FsWrapper) FuncFor(sym string) (resource.EntryFunc, error) {
	switch sym {
	case "one":
		return fs.one, nil
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
	rs := NewFsWrapper(dataDir, &st, ctx)
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
		t.Fatalf("expected where-string 'foo', got %v", r)
	}
}
