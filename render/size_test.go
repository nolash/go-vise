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

func getFoo(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	return resource.Result{
		Content: "inky",
	}, nil
}

func getBar(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	return resource.Result{
		Content: "pinky",
	}, nil
}

func getBaz(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	return resource.Result{
		Content: "blinky",
	}, nil
}

func getXyzzy(sym string, input []byte, ctx context.Context) (resource.Result, error) {
	return resource.Result{
		Content: "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve",
	}, nil
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
	err := ca.Add("foo", "inky", 4)
	if err != nil {
		t.Fatal(err)
	}
	err = ca.Add("bar", "pinky", 10)
	if err != nil {
		t.Fatal(err)
	}
	err = ca.Add("baz", "blinky", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = pg.Map("foo")
	if err != nil {
		t.Fatal(err)
	}
	err = pg.Map("bar")
	if err != nil {
		t.Fatal(err)
	}
	err = pg.Map("baz")
	if err != nil {
		t.Fatal(err)
	}

	mn.Put("1", "foo the foo")
	mn.Put("2", "go to bar")

	_, err = pg.Render("small", 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pg.Render("toobig", 0)
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

	mn.Put("1", "foo the foo")
	mn.Put("2", "go to bar")

	r, err := pg.Render("pages",  0)
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
		t.Fatalf("expected:\n\t%x\ngot:\n\t%x\n", expect, r)
	}
	r, err = pg.Render("pages", 1)
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

func TestManySizes(t *testing.T) {
	for i := 50; i < 128; i++ {
		st := state.NewState(0)
		ca := cache.NewCache()
		mn := NewMenu().WithOutputSize(32)
		mrs := resource.NewMenuResource().WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
		rs := TestSizeResource{
			mrs,	
		}
		szr := NewSizer(uint32(i))
		pg := NewPage(ca, rs).WithSizer(szr).WithMenu(mn)
		ca.Push()
		st.Down("pages")
		ca.Add("foo", "inky", 10)
		ca.Add("bar", "pinky", 10)
		ca.Add("baz", "blinky", 10)
		ca.Add("xyzzy", "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve", 0)
		pg.Map("foo")
		pg.Map("bar")
		pg.Map("baz")
		pg.Map("xyzzy")
		_, err := pg.Render("pages", 0)
		if err != nil {
			t.Fatal(err)
		}
	}
}	