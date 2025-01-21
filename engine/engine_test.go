package engine

import (
	"bytes"
	"context"
	"fmt"

	//	"io/ioutil"
	"log"
	//	"path"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/testdata"
	"git.defalsify.org/vise.git/vm"
)

var (
	dataGenerated bool   = false
	dataDir       string = testdata.DataDir
)

type testWrapper struct {
	resource.Resource
	st *state.State
	db db.Db
}

func newTestWrapper(path string, st *state.State) testWrapper {
	ctx := context.Background()
	store := fsdb.NewFsDb()
	store.Connect(ctx, path)
	rs := resource.NewDbResource(store)
	rs.With(db.DATATYPE_STATICLOAD)
	wr := testWrapper{
		rs,
		st,
		store,
	}
	rs.AddLocalFunc("one", wr.one)
	rs.AddLocalFunc("inky", wr.inky)
	rs.AddLocalFunc("pinky", wr.pinky)
	rs.AddLocalFunc("set_lang", wr.set_lang)
	rs.AddLocalFunc("translate", wr.translate)
	rs.AddLocalFunc("quit", quitFunc)
	return wr
}

func (fs testWrapper) getStore() db.Db {
	return fs.db
}

func (fs testWrapper) one(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "one",
	}, nil
}

func (fs testWrapper) inky(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "tinkywinky",
	}, nil
}

func (fs testWrapper) pinky(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	r := fmt.Sprintf("xyzzy: %x", input)
	return resource.Result{
		Content: r,
	}, nil
}

func (fs testWrapper) translate(ctx context.Context, sym string, input []byte) (resource.Result, error) {
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

func (fs testWrapper) set_lang(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: string(input),
		FlagSet: []uint32{state.FLAG_LANG},
	}, nil
}

//func(fs testWrapper) GetCode(ctx context.Context, sym string) ([]byte, error) {
//	sym += ".bin"
//	fp := path.Join(fs.Path, sym)
//	r, err := ioutil.ReadFile(fp)
//	return r, err
//}

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

func quitFunc(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "these aren't the droids you are looking for",
	}, nil
}

func TestEngineInit(t *testing.T) {
	var err error
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(17)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache().WithCacheSize(1024)

	cfg := Config{
		Root: "root",
	}
	en := NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)

	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	w := bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	b := w.Bytes()
	expect_str := `hello world
1:do the foo
2:go to the bar
3:language template`

	if !bytes.Equal(b, []byte(expect_str)) {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect_str, b)
	}

	input := []byte("1")
	_, err = en.Exec(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := st.Where()
	if r != "foo" {
		t.Fatalf("expected where-string 'foo', got %s", r)
	}
	w = bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, w)
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
	ctx := context.Background()
	st := state.NewState(17)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache().WithCacheSize(1024)

	cfg := Config{
		Root: "root",
	}
	en := NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	var err error
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = en.Exec(ctx, []byte("_foo"))
	if err == nil {
		t.Fatalf("expected fail on invalid input")
	}
}

func TestEngineResumeTerminated(t *testing.T) {
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(17)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache().WithCacheSize(1024)

	cfg := Config{
		Root: "root",
	}
	en := NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	var err error
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.Exec(ctx, []byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = en.Exec(ctx, []byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	location, idx := st.Where()
	if location != "baz" {
		t.Fatalf("expected 'baz', got %s", location)
	}
	if idx != 0 {
		t.Fatalf("expected idx '0', got %v", idx)
	}
}

func TestLanguageSet(t *testing.T) {
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(0)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache().WithCacheSize(1024)

	cfg := Config{
		Root: "root",
	}
	en := NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)

	var err error
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	b := vm.NewLine(nil, vm.LOAD, []string{"translate"}, []byte{0x01, 0xff}, nil)
	b = vm.NewLine(b, vm.LOAD, []string{"set_lang"}, []byte{0x01, 0x00}, nil)
	b = vm.NewLine(b, vm.MOVE, []string{"."}, nil, nil)
	st.SetCode(b)

	_, err = en.Exec(ctx, []byte("no"))
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

	_, err = en.Exec(ctx, []byte("no"))
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

func TestLanguageRender(t *testing.T) {
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(0)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache()

	cfg := Config{
		Root: "root",
	}
	en := NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)

	var err error
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	b := vm.NewLine(nil, vm.LOAD, []string{"set_lang"}, []byte{0x01, 0x00}, nil)
	b = vm.NewLine(b, vm.MOVE, []string{"lang"}, nil, nil)
	st.SetCode(b)

	_, err = en.Exec(ctx, []byte("nor"))
	if err != nil {
		t.Fatal(err)
	}

	br := bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, br)
	if err != nil {
		t.Fatal(err)
	}
	expect := "dette endrer"
	r := br.String()
	if r[:len(expect)] != expect {
		t.Fatalf("expected %s, got %s", expect, r[:len(expect)])
	}

}

func TestConfigLanguageRender(t *testing.T) {
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(0)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache()

	cfg := Config{
		Root:     "root",
		Language: "nor",
	}
	en := NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)

	var err error
	//_, err = en.Init(ctx)
	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	b := vm.NewLine(nil, vm.LOAD, []string{"set_lang"}, []byte{0x01, 0x00}, nil)
	b = vm.NewLine(b, vm.MOVE, []string{"lang"}, nil, nil)
	st.SetCode(b)

	_, err = en.Exec(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	br := bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, br)
	if err != nil {
		t.Fatal(err)
	}

	expect := `dette endrer med språket tinkywinky
0:tilbake`
	r := br.String()
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s", expect, r)
	}
}

func preBlock(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	log.Printf("executing preBlock")
	return resource.Result{
		Content: "None shall pass",
		FlagSet: []uint32{state.FLAG_TERMINATE},
	}, nil
}

func preAllow(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	log.Printf("executing preAllow")
	return resource.Result{}, nil
}

func TestPreVm(t *testing.T) {
	var b []byte
	var out *bytes.Buffer
	generateTestData(t)
	ctx := context.Background()
	st := state.NewState(0)
	st.UseDebug()
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache()

	cfg := Config{
		Root: "root",
	}
	en := NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	en = en.WithFirst(preBlock)
	//r, err := en.Init(ctx)
	r, err := en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if r {
		t.Fatalf("expected init to return 'not continue'")
	}
	out = bytes.NewBuffer(b)
	_, err = en.Flush(ctx, out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), []byte("None shall pass")) {
		t.Fatalf("expected writeresult 'None shall pass', got %s", out)
	}

	st = state.NewState(0)
	ca = cache.NewCache()
	en = NewEngine(cfg, &rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	en = en.WithFirst(preAllow)
	//r, err = en.Init(ctx)
	r, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Fatalf("expected init to return 'continue'")
	}
}

func TestManyQuits(t *testing.T) {
	b := bytes.NewBuffer(nil)
	ctx := context.Background()
	st := state.NewState(0)
	st.UseDebug()
	generateTestData(t)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache()

	cfg := Config{
		Root: "nothing",
	}

	en := NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	r, err := en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if r {
		t.Fatalf("expected init to return 'not continue'")
	}
	_, err = en.Flush(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	en = NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	r, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if r {
		t.Fatalf("expected init to return 'not continue'")
	}
	_, err = en.Flush(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	en = NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	r, err = en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if r {
		t.Fatalf("expected init to return 'not continue'")
	}

	b = bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	x := "these aren't the droids you are looking for"
	if !bytes.Equal(b.Bytes(), []byte(x)) {
		t.Fatalf("expected '%s', got '%s'", x, b.Bytes())
	}
}

func TestOutEmpty(t *testing.T) {
	ctx := context.Background()
	st := state.NewState(0)
	st.UseDebug()
	generateTestData(t)
	rs := newTestWrapper(dataDir, st)
	ca := cache.NewCache()

	cfg := Config{
		Root: "something",
	}

	en := NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)
	r, err := en.Exec(ctx, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	if r {
		t.Fatalf("expected init to return 'not continue'")
	}

	v := bytes.NewBuffer(nil)
	en.Flush(ctx, v)
	x := "mmmm, something..."
	if !bytes.Equal(v.Bytes(), []byte(x)) {
		t.Fatalf("expected '%s', got '%s'", x, v.Bytes())
	}
}
