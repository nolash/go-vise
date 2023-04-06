package resource

import (
	"fmt"
	"context"
	"testing"

	"git.defalsify.org/festive/state"
)

type TestSizeResource struct {
	*MenuResource
}

func getTemplate(sym string, sizer *Sizer) (string, error) {
	var tpl string
	switch sym {
	case "small":
		tpl = "one {{.foo}} two {{.bar}} three {{.baz}}"
	case "toobig":
		tpl = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus in mattis lorem. Aliquam erat volutpat. Ut vitae metus."

	}
	return tpl, nil
}

func funcFor(sym string) (EntryFunc, error) {
	switch sym {
	case "foo":
		return getFoo, nil
	case "bar":
		return getBar, nil
	case "baz":
		return getBaz, nil
	}
	return nil, fmt.Errorf("unknown func: %s", sym)
}

func getFoo(ctx context.Context) (string, error) {
	return "inky", nil
}

func getBar(ctx context.Context) (string, error) {
	return "pinky", nil
}

func getBaz(ctx context.Context) (string, error) {
	return "blinky", nil
}


func TestSizeLimit(t *testing.T) {
	st := state.NewState(0).WithOutputSize(128)
	mrs := NewMenuResource().WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
	rs := TestSizeResource{
		mrs,	
	}
	st.Add("foo", "inky", 4)
	st.Add("bar", "pinky", 10)
	st.Add("baz", "blinky", 0)
	st.Map("foo")
	st.Map("bar")
	st.Map("baz")
	st.SetMenuSize(32)
	szr, err := SizerFromState(&st)
	if err != nil {
		t.Fatal(err)
	}

	rs.PutMenu("1", "foo the foo")
	rs.PutMenu("2", "go to bar")

	tpl, err := rs.GetTemplate("small", &szr)
	if err != nil {
		t.Fatal(err)
	}

	vals, err := st.Get()
	if err != nil {
		t.Fatal(err)
	}
	_ = tpl

	_, err = rs.Render("small", vals, &szr)
	if err != nil {
		t.Fatal(err)
	}

	rs.PutMenu("1", "foo the foo")
	rs.PutMenu("2", "go to bar")

	_, err = rs.Render("toobig", vals, &szr)
	if err == nil {
		t.Fatalf("expected size exceeded")
	}
}
