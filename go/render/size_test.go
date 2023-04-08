package render

import (
	"context"
	"fmt"
	"testing"

	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/cache"
)

type TestSizeResource struct {
	*resource.MenuResource
}

func getTemplate(sym string) (string, error) {
	var tpl string
	switch sym {
	case "small":
		tpl = "one {{.foo}} two {{.bar}} three {{.baz}}"
	case "toobig":
		tpl = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus in mattis lorem. Aliquam erat volutpat. Ut vitae metus."
	case "pages":
		tpl = "one {{.foo}} two {{.bar}} three {{.baz}}\n{{.xyzzy}}"
	}
	return tpl, nil
}

func funcFor(sym string) (resource.EntryFunc, error) {
	switch sym {
	case "foo":
		return getFoo, nil
	case "bar":
		return getBar, nil
	case "baz":
		return getBaz, nil
	case "xyzzy":
		return getXyzzy, nil
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

func getXyzzy(ctx context.Context) (string, error) {
	return "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve", nil
}

func TestSizeCheck(t *testing.T) {
	szr := NewSizer(16)
	l, ok := szr.Check("foobar")
	if !ok {
		t.Fatalf("expected ok")
	}
	if l != 10 {
		t.Fatalf("expected 10, got %v", l)
	}

	l, ok = szr.Check("inkypinkyblinkyclyde")
	if ok {
		t.Fatalf("expected not ok")
	}
	if l != 0 {
		t.Fatalf("expected 0, got %v", l)
	}
}

func TestSizeLimit(t *testing.T) {
	st := state.NewState(0)
	ca := cache.NewCache()
	mn := NewMenu().WithOutputSize(32)
	mrs := resource.NewMenuResource().WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
	rs := TestSizeResource{
		mrs,
	}
	szr := NewSizer(128)
	pg := NewPage(ca, rs).WithMenu(mn).WithSizer(szr)
	ca.Push()
	st.Down("test")
	ca.Add("foo", "inky", 4)
	ca.Add("bar", "pinky", 10)
	ca.Add("baz", "blinky", 0)
	pg.Map("foo")
	pg.Map("bar")
	pg.Map("baz")

	mn.Put("1", "foo the foo")
	mn.Put("2", "go to bar")

	vals, err := ca.Get()
	if err != nil {
		t.Fatal(err)
	}

	_, err = pg.Render("small", vals, 0)
	if err != nil {
		t.Fatal(err)
	}

	mn.Put("1", "foo the foo")
	mn.Put("2", "go to bar")

	_, err = pg.Render("toobig", vals, 0)
	if err == nil {
		t.Fatalf("expected size exceeded")
	}
}

func TestSizePages(t *testing.T) {
	st := state.NewState(0)
	ca := cache.NewCache()
	mn := NewMenu().WithOutputSize(32)
	mrs := resource.NewMenuResource().WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
	rs := TestSizeResource{
		mrs,	
	}
	szr := NewSizer(128)
	pg := NewPage(ca, rs).WithSizer(szr).WithMenu(mn)
	ca.Push()
	st.Down("test")
	ca.Add("foo", "inky", 4)
	ca.Add("bar", "pinky", 10)
	ca.Add("baz", "blinky", 20)
	ca.Add("xyzzy", "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve", 0)
	pg.Map("foo")
	pg.Map("bar")
	pg.Map("baz")
	pg.Map("xyzzy")

	vals, err := ca.Get()
	if err != nil {
		t.Fatal(err)
	}

	mn.Put("1", "foo the foo")
	mn.Put("2", "go to bar")

	r, err := pg.Render("pages", vals, 0)
	if err != nil {
		t.Fatal(err)
	}

	expect := `one inky two pinky three blinky
inky pinky
blinky clyde sue
tinkywinky dipsy
lala poo
1:foo the foo
2:go to bar`


	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}
	r, err = pg.Render("pages", vals, 1)
	if err != nil {
		t.Fatal(err)
	}

	expect = `one inky two pinky three blinky
one two three four five six seven
eight nine ten
eleven twelve
1:foo the foo
2:go to bar`
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}

}
