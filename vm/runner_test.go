package vm

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/internal/resourcetest"
	"git.defalsify.org/vise.git/render"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

var (
	ctx    = context.Background()
	dynVal = "three"
)

type testResource struct {
	*resourcetest.TestResource
	state        *state.State
	RootCode     []byte
	CatchContent string
}

func newTestResource(st *state.State) testResource {
	rs := resourcetest.NewTestResource()
	tr := testResource{
		TestResource: rs,
		state:        st,
	}
	rs.AddTemplate(ctx, "foo", "inky pinky blinky clyde")
	rs.AddTemplate(ctx, "bar", "inky pinky {{.one}} blinky {{.two}} clyde")
	rs.AddTemplate(ctx, "baz", "inky pinky {{.baz}} blinky clyde")
	rs.AddTemplate(ctx, "three", "{{.one}} inky pinky {{.three}} blinky clyde {{.two}}")
	rs.AddTemplate(ctx, "root", "root")
	rs.AddTemplate(ctx, "_catch", tr.CatchContent)
	rs.AddTemplate(ctx, "ouf", "ouch")
	rs.AddTemplate(ctx, "flagCatch", "flagiee")
	rs.AddMenu(ctx, "one", "one")
	rs.AddMenu(ctx, "two", "two")
	rs.AddLocalFunc("two", getTwo)
	rs.AddLocalFunc("dyn", getDyn)
	rs.AddLocalFunc("arg", tr.getInput)
	rs.AddLocalFunc("echo", getEcho)
	rs.AddLocalFunc("setFlagOne", setFlag)
	rs.AddLocalFunc("set_lang", set_lang)
	rs.AddLocalFunc("aiee", uhOh)

	var b []byte
	b = NewLine(nil, HALT, nil, nil, nil)
	rs.AddBytecode(ctx, "one", b)

	b = NewLine(nil, MOUT, []string{"repent", "0"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	rs.AddBytecode(ctx, "_catch", b)

	b = NewLine(nil, MOUT, []string{"repent", "0"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	b = NewLine(b, MOVE, []string{"_"}, nil, nil)
	rs.AddBytecode(ctx, "flagCatch", b)

	b = NewLine(nil, MOUT, []string{"oo", "1"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	rs.AddBytecode(ctx, "ouf", b)

	rs.AddBytecode(ctx, "root", tr.RootCode)

	return tr
}

func getOne(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "one",
	}, nil
}

func getTwo(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "two",
	}, nil
}

func getDyn(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: dynVal,
	}, nil
}

func getEcho(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	r := fmt.Sprintf("echo: %s", input)
	return resource.Result{
		Content: r,
	}, nil
}

func uhOh(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{}, fmt.Errorf("uh-oh spaghetti'ohs")
}

func setFlag(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	s := fmt.Sprintf("ping")
	r := resource.Result{
		Content: s,
	}
	if len(input) > 0 {
		r.FlagSet = append(r.FlagSet, uint32(input[0]))
	}
	if len(input) > 1 {
		r.FlagReset = append(r.FlagReset, uint32(input[1]))
	}
	log.Printf("setflag %v", r)
	return r, nil

}

func set_lang(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: string(input),
		FlagSet: []uint32{state.FLAG_LANG},
	}, nil
}

//
//type TestStatefulResolver struct {
//	state *state.State
//}

func (r testResource) FuncFor(ctx context.Context, sym string) (resource.EntryFunc, error) {
	switch sym {
	case "one":
		return getOne, nil
	case "two":
		return getTwo, nil
	case "dyn":
		return getDyn, nil
	case "arg":
		return r.getInput, nil
	case "echo":
		return getEcho, nil
	case "setFlagOne":
		return setFlag, nil
	case "set_lang":
		return set_lang, nil
	case "aiee":
		return uhOh, nil
	}
	return nil, fmt.Errorf("invalid function: '%s'", sym)
}

func (r testResource) getInput(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	v, err := r.state.GetInput()
	return resource.Result{
		Content: string(v),
	}, err
}

func (r testResource) getCode(sym string) ([]byte, error) {
	var b []byte
	switch sym {
	case "_catch":
		b = NewLine(b, MOUT, []string{"repent", "0"}, nil, nil)
		b = NewLine(b, HALT, nil, nil, nil)
	case "flagCatch":
		b = NewLine(b, MOUT, []string{"repent", "0"}, nil, nil)
		b = NewLine(b, HALT, nil, nil, nil)
		b = NewLine(b, MOVE, []string{"_"}, nil, nil)
	case "root":
		b = r.RootCode
	case "ouf":
		b = NewLine(b, MOUT, []string{"oo", "1"}, nil, nil)
		b = NewLine(b, HALT, nil, nil, nil)
	}

	return b, nil
}

func TestRun(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	b := NewLine(nil, MOVE, []string{"foo"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	ctx := context.Background()
	_, err := vm.Run(ctx, b)
	if err == nil {
		t.Fatalf("expected error")
	}

	b = []byte{0x01, 0x02}
	_, err = vm.Run(ctx, b)
	if err == nil {
		t.Fatalf("no error on invalid opcode")
	}
}

func TestRunLoadRender(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	st.Down("bar")

	var err error
	ctx := context.Background()
	b := NewLine(nil, LOAD, []string{"one"}, []byte{0x0a}, nil)
	b = NewLine(b, MAP, []string{"one"}, nil, nil)
	b = NewLine(b, LOAD, []string{"two"}, []byte{0x0a}, nil)
	b = NewLine(b, MAP, []string{"two"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	r, err := vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expect := "inky pinky one blinky two clyde"
	if r != expect {
		t.Fatalf("Expected\n\t%s\ngot\n\t%s\n", expect, r)
	}

	b = NewLine(nil, LOAD, []string{"two"}, []byte{0x0a}, nil)
	b = NewLine(b, MAP, []string{"two"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	b = NewLine(nil, MAP, []string{"one"}, nil, nil)
	b = NewLine(b, MAP, []string{"two"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	_, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	r, err = vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expect = "inky pinky one blinky two clyde"
	if r != expect {
		t.Fatalf("Expected %v, got %v", expect, r)
	}
}

func TestRunMultiple(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	ctx := context.Background()
	b := NewLine(nil, MOVE, []string{"test"}, nil, nil)
	b = NewLine(b, LOAD, []string{"one"}, []byte{0x00}, nil)
	b = NewLine(b, LOAD, []string{"two"}, []byte{42}, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	var err error
	b, err = vm.Run(ctx, b)
	if err == nil {
		t.Fatal(err)
	}
}

func TestRunReload(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	szr := render.NewSizer(128)
	vm := NewVm(st, &rs, ca, szr)

	ctx := context.Background()
	b := NewLine(nil, MOVE, []string{"root"}, nil, nil)
	b = NewLine(b, LOAD, []string{"dyn"}, nil, []uint8{0})
	b = NewLine(b, MAP, []string{"dyn"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	_, err := vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	r, err := vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if r != "root" {
		t.Fatalf("expected result 'root', got %v", r)
	}
	dynVal = "baz"
	b = NewLine(nil, RELOAD, []string{"dyn"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	_, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHalt(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	b := NewLine(nil, MOVE, []string{"root"}, nil, nil)
	b = NewLine(b, LOAD, []string{"one"}, nil, []uint8{0})
	b = NewLine(b, HALT, nil, nil, nil)
	b = NewLine(b, MOVE, []string{"foo"}, nil, nil)
	var err error
	ctx := context.Background()
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Error(err)
	}
	r, _ := st.Where()
	if r == "foo" {
		t.Fatalf("Expected where-symbol not to be 'foo'")
	}
	if !bytes.Equal(b[:2], []byte{0x00, MOVE}) {
		t.Fatalf("Expected MOVE instruction, found '%v'", b)
	}
}

func TestRunArg(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	input := []byte("bar")
	_ = st.SetInput(input)

	bi := NewLine(nil, INCMP, []string{"baz", "bar"}, nil, nil)
	ctx := context.Background()

	var err error
	b, err := vm.Run(ctx, bi)
	if err == nil {
		t.Fatalf("expected error")
	}
	l := len(b)
	if l != 0 {
		t.Fatalf("expected empty remainder, got length %v: %v", l, b)
	}
	r, _ := st.Where()
	if r != "baz" {
		t.Fatalf("expected where-state baz, got %v", r)
	}
}

func TestRunInputHandler(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	_ = st.SetInput([]byte("baz"))

	bi := NewLine([]byte{}, INCMP, []string{"aiee", "bar"}, nil, nil)
	bi = NewLine(bi, INCMP, []string{"foo", "baz"}, nil, nil)
	bi = NewLine(bi, LOAD, []string{"one"}, []byte{0x00}, nil)
	bi = NewLine(bi, LOAD, []string{"two"}, []byte{0x03}, nil)
	bi = NewLine(bi, MAP, []string{"one"}, nil, nil)
	bi = NewLine(bi, MAP, []string{"two"}, nil, nil)
	bi = NewLine(bi, HALT, nil, nil, nil)

	var err error
	ctx := context.Background()
	_, err = vm.Run(ctx, bi)
	if err == nil {
		t.Fatalf("expected error")
	}
	r, _ := st.Where()
	if r != "foo" {
		t.Fatalf("expected where-sym 'foo', got '%v'", r)
	}
}

func TestRunArgInvalid(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	_ = st.SetInput([]byte("foo"))

	var err error

	st.Down("root")
	b := NewLine(nil, INCMP, []string{"baz", "bar"}, nil, nil)

	ctx := context.Background()
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	location, _ := st.Where()
	if location != "_catch" {
		t.Fatalf("expected where-state _catch, got %v", location)
	}

	r, err := vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expect := `invalid input: 'foo'
0:repent`
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s", expect, r)
	}
}

func TestRunMenu(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	ctx := context.Background()

	rs.AddBytecode(ctx, "foo", []byte{})
	rs.Lock()
	b := NewLine(nil, MOVE, []string{"foo"}, nil, nil)
	b = NewLine(b, MOUT, []string{"one", "0"}, nil, nil)
	b = NewLine(b, MOUT, []string{"two", "1"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Error(err)
	}
	l := len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}

	r, err := vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expect := "inky pinky blinky clyde\n0:one\n1:two"
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}
}

func TestRunMenuBrowse(t *testing.T) {
	log.Printf("This test is incomplete, it must check the output of a menu browser once one is implemented. For now it only checks whether it can execute the runner endpoints for the instrucitons.")
	st := state.NewState(5)
	rs := newTestResource(st)
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	ctx := context.Background()

	rs.AddBytecode(ctx, "foo", []byte{})
	rs.Lock()
	b := NewLine(nil, MOVE, []string{"foo"}, nil, nil)
	b = NewLine(b, MOUT, []string{"one", "0"}, nil, nil)
	b = NewLine(b, MOUT, []string{"two", "1"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Error(err)
	}
	l := len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}

	r, err := vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expect := "inky pinky blinky clyde\n0:one\n1:two"
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}
}

func TestRunReturn(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	st.Down("root")
	b := NewLine(nil, INCMP, []string{"bar", "0"}, nil, nil)
	b = NewLine(b, INCMP, []string{"_", "1"}, nil, nil)

	ctx := context.Background()

	st.SetInput([]byte("0"))
	b, err = vm.Run(ctx, b)
	if err == nil {
		t.Fatalf("expected error")
	}
	location, _ := st.Where()
	if location != "bar" {
		t.Fatalf("expected location 'bar', got '%s'", location)
	}

	st.SetInput([]byte("1"))
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	location, _ = st.Where()
	if location != "root" {
		t.Fatalf("expected location 'root', got '%s'", location)
	}
}

func TestRunLoadInput(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	st.Down("root")
	st.SetInput([]byte("foobar"))

	b := NewLine(nil, LOAD, []string{"echo"}, []byte{0x00}, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	ctx := context.Background()

	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	r, err := ca.Get("echo")
	if err != nil {
		t.Fatal(err)
	}
	if r != "echo: foobar" {
		t.Fatalf("expected 'echo: foobar', got %s", r)
	}
}

func TestInputBranch(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, LOAD, []string{"setFlagOne"}, []byte{0x00}, nil)
	b = NewLine(b, RELOAD, []string{"setFlagOne"}, nil, nil)
	b = NewLine(b, CATCH, []string{"flagCatch"}, []byte{8}, []uint8{1})
	b = NewLine(b, CATCH, []string{"one"}, []byte{9}, []uint8{1})
	rs.RootCode = b
	rs.AddBytecode(ctx, "root", rs.RootCode)
	rs.Lock()

	ctx := context.Background()

	st.SetInput([]byte{0x08})
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	location, _ := st.Where()
	if location != "flagCatch" {
		t.Fatalf("expected 'flagCatch', got %s", location)
	}

	st.SetInput([]byte{0x09, 0x08})
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	location, _ = st.Where()
	if location != "one" {
		t.Fatalf("expected 'one', got %s", location)
	}
}

func TestInputIgnore(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, INCMP, []string{"one", "foo"}, nil, nil)
	b = NewLine(b, INCMP, []string{"two", "bar"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	ctx := context.Background()

	st.SetInput([]byte("foo"))
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	location, _ := st.Where()
	if location != "one" {
		t.Fatalf("expected 'one', got %s", location)
	}
}

func TestInputIgnoreWildcard(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, INCMP, []string{"one", "foo"}, nil, nil)
	b = NewLine(b, INCMP, []string{"two", "*"}, nil, nil)

	ctx := context.Background()

	st.SetInput([]byte("foo"))
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	location, _ := st.Where()
	if location != "one" {
		t.Fatalf("expected 'one', got %s", location)
	}
}

func TestCatchCleanMenu(t *testing.T) {
	st := state.NewState(5)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, MOUT, []string{"1", "one"}, nil, nil)
	b = NewLine(b, MOUT, []string{"2", "two"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	b = NewLine(b, INCMP, []string{"foo", "1"}, nil, nil)
	b = NewLine(b, CATCH, []string{"ouf"}, []byte{0x08}, []uint8{0x00})

	ctx := context.Background()

	st.SetInput([]byte("foo"))
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	st.SetInput([]byte("foo"))
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	_, err = vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetLang(t *testing.T) {
	st := state.NewState(0)
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	var err error

	st.Down("root")

	st.SetInput([]byte("no"))
	b := NewLine(nil, LOAD, []string{"set_lang"}, []byte{0x01, 0x00}, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	ctx := context.Background()
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	lang := *st.Language
	if lang.Code != "nor" {
		t.Fatalf("expected language 'nor',, got %s", lang.Code)
	}
}

func TestLoadError(t *testing.T) {
	st := state.NewState(0)
	st.UseDebug()
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	st.Down("root")
	st.SetInput([]byte{})
	b := NewLine(nil, LOAD, []string{"aiee"}, []byte{0x01, 0x10}, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	var err error
	ctx := context.Background()
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	r, err := vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expect := `error aiee:0
0:repent`
	if r != expect {
		t.Fatalf("expected: \n\t%s\ngot:\n\t%s", expect, r)
	}
}

func TestMatchFlag(t *testing.T) {
	var err error
	ctx := context.Background()

	st := state.NewState(1)
	st.UseDebug()
	rs := newTestResource(st)
	rs.Lock()
	ca := cache.NewCache()
	vm := NewVm(st, &rs, ca, nil)

	st.Down("root")
	st.SetFlag(state.FLAG_USERSTART)
	st.SetInput([]byte{})
	b := NewLine(nil, CATCH, []string{"aiee"}, []byte{state.FLAG_USERSTART}, []uint8{1})
	b = NewLine(b, HALT, nil, nil, nil)
	b, err = vm.Run(ctx, b)
	if err == nil {
		t.Fatal(err)
	}

	st.SetFlag(state.FLAG_USERSTART)
	st.SetInput([]byte{})
	b = NewLine(nil, CATCH, []string{"aiee"}, []byte{state.FLAG_USERSTART}, []uint8{0})
	b = NewLine(b, HALT, nil, nil, nil)
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	st.Restart()
	st.SetFlag(state.FLAG_USERSTART)
	st.SetInput([]byte{})
	b = NewLine(nil, CROAK, nil, []byte{state.FLAG_USERSTART}, []uint8{1})
	b = NewLine(b, HALT, nil, nil, nil)
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	if st.MatchFlag(state.FLAG_TERMINATE, false) {
		t.Fatalf("expected terminate set")
	}

	st.Restart()
	st.SetFlag(state.FLAG_USERSTART)
	st.SetInput([]byte{})
	b = NewLine(nil, CROAK, nil, []byte{state.FLAG_USERSTART}, []uint8{0})
	b = NewLine(b, HALT, nil, nil, nil)
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	if st.MatchFlag(state.FLAG_TERMINATE, true) {
		t.Fatalf("expected no terminate")
	}
}

func TestBatchRun(t *testing.T) {
	var err error
	ctx := context.Background()
	st := state.NewState(0)
	st.Down("root")
	st.Down("one")
	st.Down("two")
	ca := cache.NewCache()
	rs := newTestResource(st)
	rs.Lock()
	b := NewLine(nil, MNEXT, []string{"fwd", "0"}, nil, nil)
	b = NewLine(b, MPREV, []string{"back", "11"}, nil, nil)
	b = NewLine(b, MSINK, nil, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	vm := NewVm(st, rs, ca, nil)

	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestErrorOut(t *testing.T) {
	var err error
	ctx := context.Background()
	st := state.NewState(0)
	ca := cache.NewCache()
	rs := newTestResource(st)
	rs.AddLocalFunc("foo", getTwo)
	rs.AddLocalFunc("aiee", uhOh)
	rs.Lock()
	b := NewLine(nil, LOAD, []string{"two"}, []byte{0x01, 0x10}, nil)
	vm := NewVm(st, rs, ca, nil)
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(vm.String(), " ok") {
		t.Fatalf("expected ok, got %s", vm.String())
	}

	st = state.NewState(0)
	b = NewLine(nil, LOAD, []string{"aiee"}, []byte{0x01, 0x10}, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	vm = NewVm(st, rs, ca, nil)
	b, err = vm.Run(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(vm.String(), ") error load: aiee") {
		t.Fatalf("expected load fail aiee in vm string, got %s", vm.String())
	}
	_, err = ca.Get("two")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ca.Get("aiee")
	if err == nil {
		t.Fatalf("expected error")
	}
}
