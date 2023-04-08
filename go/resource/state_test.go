package resource

import (
	"testing"

	"git.defalsify.org/festive/state"
)

func TestStateResourceInit(t *testing.T) {
	st := state.NewState(0)
	rs := NewMenuResource()
	_ = ToStateResource(rs).WithState(&st)
	_ = NewStateResource(&st)
}

func TestStateBrowseNoSink(t *testing.T) {
	st := state.NewState(0)
	st.Down("root")

	rs := NewStateResource(&st)
	rs.PutMenu("1", "foo")
	rs.PutMenu("2", "bar")
	err := rs.SetMenuBrowse("11", "next", false)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.SetMenuBrowse("22", "prev", true)
	if err != nil {
		t.Fatal(err)
	}
	s, err := rs.RenderMenu(0)
	if err != nil {
		t.Fatal(err)
	}

	expect := `1:foo
2:bar`
	if s != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, s)
	}
}


func TestStateBrowseSink(t *testing.T) {
	st := state.NewState(0)
	st.Down("root")

	rs := NewStateResource(&st)
	rs.PutMenu("1", "foo")
	rs.PutMenu("2", "bar")
	err := rs.SetMenuBrowse("11", "next", false)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.SetMenuBrowse("22", "prev", true)
	if err != nil {
		t.Fatal(err)
	}
	s, err := rs.RenderMenu(0)
	if err != nil {
		t.Fatal(err)
	}

	expect := `1:foo
2:bar
11:next`
	if s != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, s)
	}

	idx, err := st.Next()
	if err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Fatalf("expected idx 1, got %v", idx)
	}
	rs = NewStateResource(&st)
	rs.PutMenu("1", "foo")
	rs.PutMenu("2", "bar")
	err = rs.SetMenuBrowse("11", "next", false)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.SetMenuBrowse("22", "prev", true)
	if err != nil {
		t.Fatal(err)
	}
	s, err = rs.RenderMenu(idx)
	if err != nil {
		t.Fatal(err)
	}

	expect = `1:foo
2:bar
11:next
22:prev`
	if s != expect {
		t.Fatalf("expected:\n\t%s\ngot:\n\t%s\n", expect, s)
	}
}
