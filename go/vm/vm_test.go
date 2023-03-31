package vm

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"testing"
	"text/template"
	
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
)

var dynVal = "three"

type TestResource struct {
	state *state.State
}

func getOne(input []byte, ctx context.Context) (string, error) {
	return "one", nil
}

func getTwo(input []byte, ctx context.Context) (string, error) {
	return "two", nil
}

func getDyn(input []byte, ctx context.Context) (string, error) {
	return dynVal, nil
}

type TestStatefulResolver struct {
	state *state.State
}


func (r *TestResource) getEachArg(input []byte, ctx context.Context) (string, error) {
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
	b := []byte{0x00, 0x01, 0x03}
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
	b = NewTestOp(b, LOAD, []string{"one"}, nil, []uint8{0})
	b = NewTestOp(b, LOAD, []string{"two"}, nil, []uint8{42})
	st, _, err := Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Error(err)
	}
}

func TestRunReload(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}
	b := []byte{}
	b = NewTestOp(b, LOAD, []string{"dyn"}, nil, []uint8{0})
	b = NewTestOp(b, MAP, []string{"dyn"}, nil, nil)
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
	b = NewTestOp(b, RELOAD, []string{"dyn"}, nil, nil)
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

func NewTestOp(instructionList []byte, instruction uint16, args []string, post []byte, szPost []uint8) []byte {
	b := []byte{0x00, 0x00}
	binary.BigEndian.PutUint16(b, instruction)
	for _, arg := range args {
		b = append(b, uint8(len(arg)))
		b = append(b, []byte(arg)...)
	}
	if post != nil {
		b = append(b, uint8(len(post)))
		b = append(b, post...)
	}
	if szPost != nil {
		b = append(b, szPost...)
	}
	return append(instructionList, b...)
}
