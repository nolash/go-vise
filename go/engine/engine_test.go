package engine

import (
	"bytes"
	"context"
	"log"
	"path"
	"text/template"
	"testing"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/resource"
)

type FsWrapper struct {
	*resource.FsResource
	st state.State
}

func NewFsWrapper(path string, st state.State, ctx context.Context) FsWrapper {
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
	log.Printf("template is %v render is %v", v, b)
	return b.String(), err
}

func(fs FsWrapper) FuncFor(sym string) (resource.EntryFunc, error) {
	return nil, nil
}

func TestEngineInit(t *testing.T) {
//	cfg := Config{
//		FlagCount: 12,
//		CacheSize: 1024,
//	}	
	st := state.NewState(17).WithCacheSize(1024)
//	dir, err := ioutil.TempDir("", "festive_test_")
//	if err != nil {
//		t.Fatal(err)
//	}
	dir := path.Join(testdataloader.GetBasePath(), "testdata")
	ctx := context.TODO()
	rs := NewFsWrapper(dir, st, ctx)
	en := NewEngine(st, rs)
	err := en.Init(ctx)
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
}
