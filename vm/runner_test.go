package vm

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"
	
	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/render"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

var dynVal = "three"

type TestResource struct {
	resource.MemResource
	state *state.State
	RootCode []byte
	CatchContent string
}

func NewTestResource(st *state.State) TestResource {
	tr := TestResource{
		MemResource: resource.NewMemResource(),
		state: st,	
	}
	tr.AddTemplate("foo", "inky pinky blinky clyde")
	tr.AddTemplate("bar", "inky pinky {{.one}} blinky {{.two}} clyde")
	tr.AddTemplate("baz", "inky pinky {{.baz}} blinky clyde")
	tr.AddTemplate("three", "{{.one}} inky pinky {{.three}} blinky clyde {{.two}}")
	tr.AddTemplate("root", "root")
	tr.AddTemplate("_catch", tr.CatchContent)
	tr.AddTemplate("ouf", "ouch")
	tr.AddTemplate("flagCatch", "flagiee")
	tr.AddEntryFunc("one", getOne)
	tr.AddEntryFunc("two", getTwo)
	tr.AddEntryFunc("dyn", getDyn)
	tr.AddEntryFunc("arg", tr.getInput)
	tr.AddEntryFunc("echo", getEcho)
	tr.AddEntryFunc("setFlagOne", setFlag)
	tr.AddEntryFunc("set_lang", set_lang)
	tr.AddEntryFunc("aiee", uhOh)

	var b []byte
	b = NewLine(nil, HALT, nil, nil, nil)
	tr.AddBytecode("one", b)

	b = NewLine(nil, MOUT, []string{"repent", "0"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	tr.AddBytecode("_catch", b)

	b = NewLine(nil, MOUT, []string{"repent", "0"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	b = NewLine(b, MOVE, []string{"_"}, nil, nil)
	tr.AddBytecode("flagCatch", b)

	b = NewLine(nil, MOUT, []string{"oo", "1"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	tr.AddBytecode("ouf", b)
	
	tr.AddBytecode("root", tr.RootCode)
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

func (r TestResource) FuncFor(sym string) (resource.EntryFunc, error) {
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

func(r TestResource) getInput(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	v, err := r.state.GetInput()
	return resource.Result{
		Content: string(v),
	}, err
}

func(r TestResource) getCode(sym string) ([]byte, error) {
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	b := NewLine(nil, MOVE, []string{"foo"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	st.Down("bar")

	var err error
	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	szr := render.NewSizer(128)
	vm := NewVm(&st, &rs, ca, szr)

	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	b := NewLine(nil, MOVE, []string{"root"}, nil, nil)
	b = NewLine(b, LOAD, []string{"one"}, nil, []uint8{0})
	b = NewLine(b, HALT, nil, nil, nil)
	b = NewLine(b, MOVE, []string{"foo"}, nil, nil)
	var err error
	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	input := []byte("bar")
	_ = st.SetInput(input)

	bi := NewLine(nil, INCMP, []string{"baz", "bar"}, nil, nil)
	ctx := context.TODO()

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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	_ = st.SetInput([]byte("baz"))

	bi := NewLine([]byte{}, INCMP, []string{"aiee", "bar"}, nil, nil)
	bi = NewLine(bi, INCMP, []string{"foo", "baz"}, nil, nil)
	bi = NewLine(bi, LOAD, []string{"one"}, []byte{0x00}, nil)
	bi = NewLine(bi, LOAD, []string{"two"}, []byte{0x03}, nil)
	bi = NewLine(bi, MAP, []string{"one"}, nil, nil)
	bi = NewLine(bi, MAP, []string{"two"}, nil, nil)
	bi = NewLine(bi, HALT, nil, nil, nil)

	var err error
	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	_ = st.SetInput([]byte("foo"))

	var err error
	
	st.Down("root")
	b := NewLine(nil, INCMP, []string{"baz", "bar"}, nil, nil)

	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	ctx := context.TODO()

	rs.AddBytecode("foo", []byte{})
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	ctx := context.TODO()

	rs.AddBytecode("foo", []byte{})
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	st.Down("root")
	b := NewLine(nil, INCMP, []string{"bar", "0"}, nil, nil)
	b = NewLine(b, INCMP, []string{"_", "1"}, nil, nil)

	ctx := context.TODO()

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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	st.Down("root")
	st.SetInput([]byte("foobar"))

	b := NewLine(nil, LOAD, []string{"echo"}, []byte{0x00}, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	ctx := context.TODO()

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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, LOAD, []string{"setFlagOne"}, []byte{0x00}, nil)
	b = NewLine(b, RELOAD, []string{"setFlagOne"}, nil, nil)
	b = NewLine(b, CATCH, []string{"flagCatch"}, []byte{8}, []uint8{1})
	b = NewLine(b, CATCH, []string{"one"}, []byte{9}, []uint8{1})
	rs.RootCode = b
	rs.AddBytecode("root", rs.RootCode)

	ctx := context.TODO()

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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, INCMP, []string{"one", "foo"}, nil, nil)
	b = NewLine(b, INCMP, []string{"two", "bar"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	ctx := context.TODO()

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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, INCMP, []string{"one", "foo"}, nil, nil)
	b = NewLine(b, INCMP, []string{"two", "*"}, nil, nil)

	ctx := context.TODO()

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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	st.Down("root")

	b := NewLine(nil, MOUT, []string{"1", "one"}, nil, nil)
	b = NewLine(b, MOUT, []string{"2", "two"}, nil, nil)
	b = NewLine(b, HALT, nil, nil, nil)
	b = NewLine(b, INCMP, []string{"foo", "1"}, nil, nil)
	b = NewLine(b, CATCH, []string{"ouf"}, []byte{0x08}, []uint8{0x00})

	ctx := context.TODO()

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

	r, err := vm.Render(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Result:\n%s", r)
}

func TestSetLang(t *testing.T) {
	st := state.NewState(0)
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	var err error

	st.Down("root")

	st.SetInput([]byte("no"))
	b := NewLine(nil, LOAD, []string{"set_lang"}, []byte{0x01, 0x00}, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	ctx := context.TODO()
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
	rs := NewTestResource(&st)
	ca := cache.NewCache()
	vm := NewVm(&st, &rs, ca, nil)

	st.Down("root")
	st.SetInput([]byte{})
	b := NewLine(nil, LOAD, []string{"aiee"}, []byte{0x01, 0x10}, nil)
	b = NewLine(b, HALT, nil, nil, nil)

	var err error
	ctx := context.TODO()
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
