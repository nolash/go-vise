package engine

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/testdata"
	"git.defalsify.org/vise.git/vm"
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
	wr.AddLocalFunc("set_lang", wr.set_lang)
	wr.AddLocalFunc("translate", wr.translate)
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

func(fs FsWrapper) translate(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	r := "cool"
	v := ctx.Value("Language")
	code := ""
	lang, ok := v.(lang.Language)
	if ok {
		code = lang.Code
	}
	if code == "nor" {
		r = "fett"
	}
	return resource.Result{
		Content: r,
	}, nil
}

func(fs FsWrapper) set_lang(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	return resource.Result{
		Content: string(input),
		FlagSet: []uint32{state.FLAG_LANG},
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

	_, err = en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
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

	
	en := NewEngine(Config{
		Root: "root",
	}, &st, &rs, ca, ctx)
	var err error
	_, err = en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = en.Exec([]byte("_foo"), ctx)
	if err == nil {
		t.Fatalf("expected fail on invalid input")
	}
}

func TestEngineResumeTerminated(t *testing.T) {
	generateTestData(t)
	ctx := context.TODO()
	st := state.NewState(17)
	rs := NewFsWrapper(dataDir, &st)
	ca := cache.NewCache().WithCacheSize(1024)

	en := NewEngine(Config{
		Root: "root",
	}, &st, &rs, ca, ctx)
	var err error
	_, err = en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.Exec([]byte("1"), ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.Exec([]byte("1"), ctx)
	if err != nil {
		t.Fatal(err)
	}

	location, idx := st.Where()
	if location != "root" {
		t.Fatalf("expected 'root', got %s", location)
	}
	if idx != 0 {
		t.Fatalf("expected idx '0', got %v", idx)
	}
}

func TestLanguageSet(t *testing.T) {
	generateTestData(t)
	ctx := context.TODO()
	st := state.NewState(0)
	rs := NewFsWrapper(dataDir, &st)
	ca := cache.NewCache().WithCacheSize(1024)

	en := NewEngine(Config{
		Root: "root",
	}, &st, &rs, ca, ctx)

	var err error
	_, err = en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}

	b := vm.NewLine(nil, vm.LOAD, []string{"translate"}, []byte{0x01, 0xff}, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"set_lang"}, []byte{0x01, 0x00}, nil)
	b = vm.NewLine(b, vm.MOVE, []string{"."}, nil, nil)
	st.SetCode(b)

	_, err = en.Exec([]byte("no"), ctx)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ca.Get("translate")
	if err != nil {
		t.Fatal(err)
	}
	if r != "cool" {
		t.Fatalf("expected 'cool', got '%s'", r)
	}


	b = vm.NewLine(nil, vm.RELOAD, []string{"translate"}, nil, nil)
	b = vm.NewLine(b, vm.MOVE, []string{"."}, nil, nil)
	st.SetCode(b)

	_, err = en.Exec([]byte("no"), ctx)
	if err != nil {
		t.Fatal(err)
	}
	r, err = ca.Get("translate")
	if err != nil {
		t.Fatal(err)
	}
	if r != "fett" {
		t.Fatalf("expected 'fett', got '%s'", r)
	}
}
