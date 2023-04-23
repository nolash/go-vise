package render

import (
	"fmt"
	"testing"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/resource"
)


func TestPageCurrentSize(t *testing.T) {
	t.Skip("usage is not in use, and it is unclear how it should be calculated")
	ca := cache.NewCache()
	pg := NewPage(ca, nil)
	err := ca.Push()
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("foo", "inky", 0)
	if err != nil {
		t.Error(err)
	}
	err = ca.Push()
	pg.Reset()
	err = ca.Add("bar", "pinky", 10)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("baz", "tinkywinkydipsylalapoo", 51)
	if err != nil {
		t.Error(err)
	}
	err = pg.Map("foo")
	if err != nil {
		t.Error(err)
	}
	err = pg.Map("bar")
	if err != nil {
		t.Error(err)
	}
	err = pg.Map("baz")
	if err != nil {
		t.Error(err)
	}
	l, c, err := pg.Usage()
	if err != nil {
		t.Error(err)
	}
	if l != 27 {
		t.Errorf("expected actual length 27, got %v", l)
	}
	if c != 34 {
		t.Errorf("expected remaining length 34, got %v", c)
	}

	mn := NewMenu().WithOutputSize(32)
	pg = pg.WithMenu(mn)
	l, c, err = pg.Usage()
	if err != nil {
		t.Error(err)
	}
	if l != 59 {
		t.Errorf("expected actual length 59, got %v", l)
	}
	if c != 2 {
		t.Errorf("expected remaining length 2, got %v", c)
	}
}

func TestStateMapSink(t *testing.T) {
	ca := cache.NewCache()
	pg := NewPage(ca, nil)
	ca.Push()
	err := ca.Add("foo", "bar", 0)
	if err != nil {
		t.Error(err)
	}
	ca.Push()
	pg.Reset()
	err = ca.Add("bar", "xyzzy", 6)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("baz", "bazbaz", 18)
	if err != nil {
		t.Error(err)
	}
	err = ca.Add("xyzzy", "plugh", 0)
	if err != nil {
		t.Error(err)
	}
	err = pg.Map("foo")
	if err != nil {
		t.Error(err)
	}
	err = pg.Map("xyzzy")
	if err == nil {
		t.Errorf("Expected fail on duplicate sink")
	}
	err = pg.Map("baz")
	if err != nil {
		t.Error(err)
	}
	ca.Push()
	pg.Reset()
	err = pg.Map("foo")
	if err != nil {
		t.Error(err)
	}
	ca.Pop()
	pg.Reset()
	err = pg.Map("foo")
	if err != nil {
		t.Error(err)
	}
}

func TestWithError(t *testing.T) {
	ca := cache.NewCache()
	rs := resource.NewMemResource()
	pg := NewPage(ca, rs)
	ca.Push()

	mn := NewMenu().WithOutputSize(32)
	err := mn.Put("0", "aiee")
	if err != nil {
		t.Fatal(err)
	}
	err = fmt.Errorf("my humps")
	pg = pg.WithMenu(mn).WithError(err)

}
