package resource

import (
	"testing"

	"git.defalsify.org/festive/state"
)

func TestSizeLimit(t *testing.T) {
	st := state.NewState(0).WithOutputSize(128)
	mrs := NewMenuResource().WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
	rs := TestSizeResource{
		mrs,	
	}
	st.Down("test")
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

	_, err = rs.Render("small", vals, 0, &szr)
	if err != nil {
		t.Fatal(err)
	}

	rs.PutMenu("1", "foo the foo")
	rs.PutMenu("2", "go to bar")

	_, err = rs.Render("toobig", vals, 0, &szr)
	if err == nil {
		t.Fatalf("expected size exceeded")
	}
}

func TestSizePages(t *testing.T) {
	st := state.NewState(0).WithOutputSize(128)
	mrs := NewMenuResource().WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
	rs := TestSizeResource{
		mrs,	
	}
	st.Down("test")
	st.Add("foo", "inky", 4)
	st.Add("bar", "pinky", 10)
	st.Add("baz", "blinky", 20)
	st.Add("xyzzy", "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve", 0)
	st.Map("foo")
	st.Map("bar")
	st.Map("baz")
	st.Map("xyzzy")
	st.SetMenuSize(32)
	szr, err := SizerFromState(&st)
	if err != nil {
		t.Fatal(err)
	}

	vals, err := st.Get()
	if err != nil {
		t.Fatal(err)
	}

	rs.PutMenu("1", "foo the foo")
	rs.PutMenu("2", "go to bar")

	r, err := rs.Render("pages", vals, 0, &szr)
	if err != nil {
		t.Fatal(err)
	}

	expect := `one inky two pinky three blinky
inky pinky
blinky clyde sue
tinkywinky dipsy
lala poo`


	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}

	rs.PutMenu("1", "foo the foo")
	rs.PutMenu("2", "go to bar")

	szr, err = SizerFromState(&st)
	if err != nil {
		t.Fatal(err)
	}
	r, err = rs.Render("pages", vals, 1, &szr)
	if err != nil {
		t.Fatal(err)
	}

	expect = `one inky two pinky three blinky
one two three four five six seven
eight nine ten
eleven twelve`
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}

}
