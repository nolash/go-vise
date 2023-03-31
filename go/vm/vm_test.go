package vm

import (
	"context"
	"fmt"
	"testing"
	"text/template"
	"bytes"
	
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
)

type TestResource struct {
}

func getOne(input []byte, ctx context.Context) (string, error) {
	return "one", nil
}

func getTwo(input []byte, ctx context.Context) (string, error) {
	return "two", nil
}

func (r *TestResource) Get(sym string) (string, error) {
	switch sym {
	case "foo":
		return "inky pinky blinky clyde", nil
	case "bar":
		return "inky pinky {{.one}} blinky {{.two}} clyde", nil
	}
	return "", fmt.Errorf("unknown symbol %s", sym)
}

func (r *TestResource) Render(sym string, values map[string]string) (string, error) {
	v, err := r.Get(sym)
	if err != nil {
		return "", err
	}
	t, err := template.New("tester").Option("missingkey=error").Parse(v)
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer([]byte{})
	err = t.Execute(b, values)
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
	}
	return nil, fmt.Errorf("invalid function: '%s'", sym)
}

func TestRun(t *testing.T) {
	st := state.NewState(5)
	rs := TestResource{}
	b := []byte{0x00, 0x02}
	r, err := Run(b, st, &rs, context.TODO())
	if err != nil {
		t.Errorf("error on valid opcode: %v", err)	
	}

	b = []byte{0x01, 0x02}
	r, err = Run(b, st, &rs, context.TODO())
	if err == nil {
		t.Errorf("no error on invalid opcode")	
	}
	_ = r
}

func TestRunMap(t *testing.T) {
	st := state.NewState(5)
	st.Enter("barbarbar")
	rs := TestResource{}
	sym := "one"
	ins := append([]byte{uint8(len(sym))}, []byte(sym)...)
	var err error
	st, err = RunMap(ins, st, &rs, context.TODO())
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
	st, err = RunMap(ins, st, &rs, context.TODO())
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
