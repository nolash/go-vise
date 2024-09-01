package render

import (
	"context"
	"fmt"
	"testing"

	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/internal/resourcetest"
	"git.defalsify.org/vise.git/cache"
)

type testSizeResource struct {
	*resourcetest.TestResource
}

func newTestSizeResource() *testSizeResource {
	ctx := context.Background()
	rs := resourcetest.NewTestResource()
	tr := &testSizeResource{
		TestResource: rs,
	}
	rs.AddTemplate(ctx, "small", "one {{.foo}} two {{.bar}} three {{.baz}}")
	rs.AddTemplate(ctx, "toobug", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus in mattis lorem. Aliquam erat volutpat. Ut vitae metus.")
	rs.AddTemplate(ctx, "pages", "one {{.foo}} two {{.bar}} three {{.baz}}\n{{.xyzzy}}")
	rs.AddLocalFunc("foo", get)
	rs.AddLocalFunc("bar", get)
	rs.AddLocalFunc("baz", get)
	rs.AddLocalFunc("xyzzy", getXyzzy)
	return tr
}

func get(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	switch sym {
	case "foo":
		return resource.Result{
			Content: "inky",
		}, nil
	case "bar":
		return resource.Result{
			Content: "pinky",
		}, nil
	case "baz":
		return resource.Result{
			Content: "blinky",
		}, nil
	}
	return resource.Result{}, fmt.Errorf("unknown sym: %s", sym)
}

func getXyzzy(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	r := "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve"
	return resource.Result{
		Content: r,
	}, nil
}
//
//type TestSizeResource struct {
//	*resource.MenuResource
//}
//
//func getTemplate(sym string, ctx context.Context) (string, error) {
//	var tpl string
//	switch sym {
//	case "small":
//		tpl = "one {{.foo}} two {{.bar}} three {{.baz}}"
//	case "toobig":
//		tpl = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus in mattis lorem. Aliquam erat volutpat. Ut vitae metus."
//	case "pages":
//		tpl = "one {{.foo}} two {{.bar}} three {{.baz}}\n{{.xyzzy}}"
//	}
//	return tpl, nil
//}
//
//func funcFor(sym string) (resource.EntryFunc, error) {
//	switch sym {
//	case "foo":
//		return getFoo, nil
//	case "bar":
//		return getBar, nil
//	case "baz":
//		return getBaz, nil
//	case "xyzzy":
//		return getXyzzy, nil
//	}
//	return nil, fmt.Errorf("unknown func: %s", sym)
//}
//
//func getFoo(sym string, input []byte, ctx context.Context) (resource.Result, error) {
//	return resource.Result{
//		Content: "inky",
//	}, nil
//}
//
//func getBar(sym string, input []byte, ctx context.Context) (resource.Result, error) {
//	return resource.Result{
//		Content: "pinky",
//	}, nil
//}
//
//func getBaz(sym string, input []byte, ctx context.Context) (resource.Result, error) {
//	return resource.Result{
//		Content: "blinky",
//	}, nil
//}
//
//func getXyzzy(sym string, input []byte, ctx context.Context) (resource.Result, error) {
//	return resource.Result{
//		Content: "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve",
//	}, nil
//}

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
	mn := NewMenu() //.WithOutputSize(32)
	//mrs := NewMenuResource() //.WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
	//rs := TestSizeResource{
	//	mrs,
	//}
	rs := newTestSizeResource()
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

	ctx := context.Background()
	_, err = pg.Render(ctx, "small", 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pg.Render(ctx, "toobig", 0)
	if err == nil {
		t.Fatalf("expected size exceeded")
	}
}

func TestSizePages(t *testing.T) {
	st := state.NewState(0)
	ca := cache.NewCache()
	mn := NewMenu() //.WithOutputSize(32)
	//mrs := NewMenuResource() //.WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
	//rs := TestSizeResource{
	//	mrs,	
	//}
	rs := newTestSizeResource()
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

	ctx := context.Background()
	r, err := pg.Render(ctx, "pages",  0)
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
	r, err = pg.Render(ctx, "pages", 1)
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
	for i := 60; i < 160; i++ {
		st := state.NewState(0)
		ca := cache.NewCache()
		mn := NewMenu() //.WithOutputSize(32)
		rs := newTestSizeResource() //.WithEntryFuncGetter(funcFor).WithTemplateGetter(getTemplate)
		//rs := TestSizeResource{
		//	mrs,	
		//}
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

		ctx := context.Background()
		_, err := pg.Render(ctx, "pages", 0)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestManySizesMenued(t *testing.T) {
	for i := 60; i < 160; i++ {
		st := state.NewState(0)
		ca := cache.NewCache()
		mn := NewMenu() //.WithOutputSize(32)
		rs := newTestSizeResource()
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
		mn.Put("0", "yay")
		mn.Put("12", "nay")

		ctx := context.Background()
		_, err := pg.Render(ctx, "pages", 0)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestMenuCollideSink(t *testing.T) {
	ctx := context.Background()
	ca := cache.NewCache()
	rs := resourcetest.NewTestResource()
	rs.AddTemplate(ctx, "foo", "bar")
	szr := NewSizer(30)
	pg := NewPage(ca, rs).WithSizer(szr)
	ca.Push()

	ca.Add("inky", "pinky", 5)
	ca.Add("blinky", "clyde", 0)
	pg.Map("inky")
	
	mn := NewMenu().WithSink()
	pg = pg.WithMenu(mn)

	var err error
	_, err = pg.Render(ctx, "foo", 0)
	if err != nil {
		t.Fatal(err)
	}
	
	mn = NewMenu().WithSink()
	pg = pg.WithMenu(mn)
	pg.Map("blinky")
	_, err = pg.Render(ctx, "foo", 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestMenuSink(t *testing.T) {
	var err error
	ctx := context.Background()

	ca := cache.NewCache()
	rs := resourcetest.NewTestResource()
	rs.AddTemplate(ctx, "foo", "bar {{.baz}}")
	szr := NewSizer(45)

	mn := NewMenu().WithSink().WithBrowseConfig(DefaultBrowseConfig())
	mn.Put("0", "inky")
	mn.Put("1", "pinky")
	mn.Put("22", "blinky")
	mn.Put("3", "clyde")
	mn.Put("44", "tinkywinky")

	pg := NewPage(ca, rs).WithSizer(szr).WithMenu(mn)
	ca.Push()

	ca.Add("baz", "xyzzy", 5)
	pg.Map("baz")

	r, err := pg.Render(ctx, "foo", 0)
	if err != nil {
		t.Fatal(err)
	}
	expect := `bar xyzzy
0:inky
1:pinky
22:blinky
11:next`
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}

	mn = NewMenu().WithSink().WithBrowseConfig(DefaultBrowseConfig())
	mn.Put("0", "inky")
	mn.Put("1", "pinky")
	mn.Put("22", "blinky")
	mn.Put("3", "clyde")
	mn.Put("44", "tinkywinky")

	pg = NewPage(ca, rs).WithSizer(szr).WithMenu(mn)
	ca.Push()

	ca.Add("baz", "xyzzy", 5)
	pg.Map("baz")

	r, err = pg.Render(ctx, "foo", 1)
	if err != nil {
		t.Fatal(err)
	}
	expect = `bar xyzzy
3:clyde
44:tinkywinky
22:previous`
	if r != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, r)
	}
}


