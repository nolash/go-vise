package vm

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"
	"text/template"
	
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/router"
	"git.defalsify.org/festive/state"
)

var dynVal = "three"

type TestResource struct {
	state *state.State
}

func getOne(ctx context.Context) (string, error) {
	return "one", nil
}

func getTwo(ctx context.Context) (string, error) {
	return "two", nil
}

func getDyn(ctx context.Context) (string, error) {
	return dynVal, nil
}

type TestStatefulResolver struct {
	state *state.State
}


func (r *TestResource) getEachArg(ctx context.Context) (string, error) {
	return r.state.PopArg()
}

func (r *TestResource) Get(sym string) (string, error) {
	switch sym {
	case "foo":
		return "inky pinky blinky clyde", nil
	case "bar":
		return "inky pinky {{.one}} blinky {{.two}} clyde", nil
	case "baz":
		return "inky pinky {{.baz}} blinky clyde", nil
	case "three":
		return "{{.one}} inky pinky {{.three}} blinky clyde {{.two}}", nil
	}
	panic(fmt.Sprintf("unknown symbol %s", sym))
	return "", fmt.Errorf("unknown symbol %s", sym)
}

func (r *TestResource) Render(sym string, values map[string]string) (string, error) {
	v, err := r.Get(sym)
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

func (r *TestResource) FuncFor(sym string) (resource.EntryFunc, error) {
	switch sym {
	case "one":
		return getOne, nil
	case "two":
		return getTwo, nil
	case "dyn":
		return getDyn, nil
	case "arg":
		return r.getEachArg, nil
	}
	return nil, fmt.Errorf("invalid function: '%s'", sym)
}

func TestRun(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}
	b := []byte{0x00, MOVE, 0x03}
	b = append(b, []byte("foo")...)
	r, _, err := Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Errorf("error on valid opcode: %v", err)	
	}

	b = []byte{0x01, 0x02}
	r, _, err = Run(b, st, &rs, context.TODO())
	if err == nil {
		t.Errorf("no error on invalid opcode")	
	}
	_ = r
}

func TestRunLoadRender(t *testing.T) {
	st := state.NewState(5)
	st.Down("barbarbar")
	rs := TestResource{}
	sym := "one"
	ins := append([]byte{uint8(len(sym))}, []byte(sym)...)
	ins = append(ins, 0x0a)
	var err error
	st, _, err = RunLoad(ins, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	m, err := st.Get()
	if err != nil {
		t.Error(err)
	}
	r, err := rs.Render("foo", m)
	if err != nil {
		t.Error(err)
	}
	expect := "inky pinky blinky clyde"
	if r != expect {
		t.Errorf("Expected %v, got %v", []byte(expect), []byte(r))
	}

	r, err = rs.Render("bar", m)
	if err == nil {
		t.Errorf("expected error for render of bar: %v" ,err)
	}

	sym = "two"
	ins = append([]byte{uint8(len(sym))}, []byte(sym)...)
	ins = append(ins, 0)
	st, _, err = RunLoad(ins, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	m, err = st.Get()
	if err != nil {
		t.Error(err)
	}
	r, err = rs.Render("bar", m)
	if err != nil {
		t.Error(err)
	}
	expect = "inky pinky one blinky two clyde"
	if r != expect {
		t.Errorf("Expected %v, got %v", expect, r)
	}
}

func TestRunMultiple(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}
	b := []byte{}
	b = NewLine(b, LOAD, []string{"one"}, nil, []uint8{0})
	b = NewLine(b, LOAD, []string{"two"}, nil, []uint8{42})
	st, _, err := Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
}

func TestRunReload(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}
	b := []byte{}
	b = NewLine(b, LOAD, []string{"dyn"}, nil, []uint8{0})
	b = NewLine(b, MAP, []string{"dyn"}, nil, nil)
	st, _, err := Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	r, err := st.Val("dyn")
	if err != nil {
		t.Error(err)
	}
	if r != "three" {
		t.Errorf("expected result 'three', got %v", r)
	}
	dynVal = "baz"
	b = []byte{}
	b = NewLine(b, RELOAD, []string{"dyn"}, nil, nil)
	st, _, err = Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	r, err = st.Val("dyn")
	if err != nil {
		t.Error(err)
	}
	log.Printf("dun now %s", r)
	if r != "baz" {
		t.Errorf("expected result 'baz', got %v", r)
	}

}

func TestRunArg(t *testing.T) {
	st := state.NewState(5)
	rt := router.NewRouter()
	rt.Add("foo", "bar")
	rt.Add("baz", "xyzzy")
	b := []byte{0x03}
	b = append(b, []byte("baz")...)
	b = append(b, rt.ToBytes()...)
	var err error
	st, b, err = Apply(b, []byte{}, st, nil, context.TODO())
	if err != nil {
		t.Error(err)	
	}
	l := len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}
	r := st.Where()
	if r != "xyzzy" {
		t.Errorf("expected where-state baz, got %v", r)
	}
}

func TestRunArgInvalid(t *testing.T) {
	st := state.NewState(5)
	rt := router.NewRouter()
	rt.Add("foo", "bar")
	rt.Add("baz", "xyzzy")
	b := []byte{0x03}
	b = append(b, []byte("bar")...)
	b = append(b, rt.ToBytes()...)
	var err error
	st, b, err = Apply(b, []byte{}, st, nil, context.TODO())
	if err != nil {
		t.Error(err)	
	}
	l := len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}
	r := st.Where()
	if r != "_catch" {
		t.Errorf("expected where-state _catch, got %v", r)
	}
}

func TestRunArgInstructions(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}

	rt := router.NewRouter()
	rt.Add("foo", "bar")
	b := []byte{0x03}
	b = append(b, []byte("foo")...)
	b = append(b, rt.ToBytes()...)

	bi := NewLine([]byte{}, LOAD, []string{"one"}, nil, []uint8{0})
	bi = NewLine(bi, LOAD, []string{"two"}, nil, []uint8{3})
	bi = NewLine(bi, MAP, []string{"one"}, nil, nil)
	bi = NewLine(bi, MAP, []string{"two"}, nil, nil)
	var err error
	st, b, err = Apply(b, bi, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)	
	}
	l := len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}
	loc := st.Where()
	if loc != "bar" {
		t.Errorf("expected where-state _catch, got %v", loc)
	}
	m, err := st.Get()
	if err != nil {
		t.Error(err)	
	}
	r, err := rs.Render(loc, m)
	if err != nil {
		t.Error(err)	
	}
	_ = r
}

func TestRunMoveAndBack(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}
	rt := router.NewRouter()
	rt.Add("foo", "bar")
	b := []byte{0x03}
	b = append(b, []byte("foo")...)
	b = append(b, rt.ToBytes()...)
	bi := NewLine([]byte{}, LOAD, []string{"one"}, nil, []uint8{0})

	var err error
	st, b, err = Apply(b, bi, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	l := len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}

	rt = router.NewRouter()
	rt.Add("foo", "baz")
	b = []byte{0x03}
	b = append(b, []byte("foo")...)
	b = append(b, rt.ToBytes()...)
	bi = NewLine([]byte{}, LOAD, []string{"two"}, nil, []uint8{0})
	st, b, err = Apply(b, bi, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	l = len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}

	rt = router.NewRouter()
	rt.Add("foo", "_")
	b = []byte{0x03}
	b = append(b, []byte("foo")...)
	b = append(b, rt.ToBytes()...)
	st, b, err = Apply(b, []byte{}, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	l = len(b)
	if l != 0 {
		t.Errorf("expected empty remainder, got length %v: %v", l, b)
	}
	loc := st.Where()
	if loc != "bar" {
		t.Errorf("expected where-string 'bar', got %v", loc)
	}
}

func TestCatchAndBack(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}
	rt := router.NewRouter()
	rt.Add("foo", "bar")
	b := NewLine([]byte{}, LOAD, []string{"one"}, nil, []uint8{0})
	b = NewLine(b, CATCH, []string{"bar"}, []byte{0x04}, nil)
	b = NewLine(b, MOVE, []string{"foo"}, nil, nil)
	st, _, err := Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	r := st.Where()
	if r != "foo" {
		t.Errorf("expected where-symbol 'foo', got %v", r)
	}

	st.SetFlag(2)
	b = NewLine([]byte{}, LOAD, []string{"two"}, nil, []uint8{0})
	b = NewLine(b, CATCH, []string{"bar"}, []byte{0x04}, nil)
	b = NewLine(b, MOVE, []string{"foo"}, nil, nil)
	st, _, err = Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
	r = st.Where()
	if r != "bar" {
		t.Errorf("expected where-symbol 'bar', got %v", r)
	}
}
