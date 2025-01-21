package resource

import (
	"bytes"
	"context"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/db/mem"
)

func TestDb(t *testing.T) {
	var rsifc Resource
	ctx := context.Background()
	store := mem.NewMemDb()
	store.Connect(ctx, "")
	tg := NewDbResource(store)
	tg.Without(db.DATATYPE_BIN)
	tg.Without(db.DATATYPE_MENU)
	tg.Without(db.DATATYPE_TEMPLATE)
	// check that it fulfills interface
	rsifc = tg
	_ = rsifc
	rs := NewMenuResource()
	rs.WithTemplateGetter(tg.GetTemplate)

	s, err := rs.GetTemplate(ctx, "foo")
	if err == nil {
		t.Fatal("expected error")
	}

	store.SetPrefix(db.DATATYPE_TEMPLATE)
	err = store.Put(ctx, []byte("foo"), []byte("bar"))
	if err == nil {
		t.Fatal("expected error")
	}
	store.SetLock(db.DATATYPE_TEMPLATE, false)
	err = store.Put(ctx, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	store.SetLock(db.DATATYPE_TEMPLATE, true)
	tg.With(db.DATATYPE_TEMPLATE)
	s, err = rs.GetTemplate(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if s != "bar" {
		t.Fatalf("expected 'bar', got %s", s)
	}

	// test support check
	store.SetPrefix(db.DATATYPE_BIN)
	store.SetLock(db.DATATYPE_BIN, false)
	err = store.Put(ctx, []byte("xyzzy"), []byte("deadbeef"))
	if err != nil {
		t.Fatal(err)
	}
	store.SetLock(db.DATATYPE_BIN, true)

	rs.WithCodeGetter(tg.GetCode)
	b, err := rs.GetCode(ctx, "xyzzy")
	if err == nil {
		t.Fatalf("expected error")
	}
	tg.With(db.DATATYPE_BIN)
	b, err = rs.GetCode(ctx, "xyzzy")
	if err != nil {
		t.Fatal(err)
	}

	tg = NewDbResource(store)
	rs.WithTemplateGetter(tg.GetTemplate)

	rs.WithCodeGetter(tg.GetCode)
	b, err = rs.GetCode(ctx, "xyzzy")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("deadbeef")) {
		t.Fatalf("expected 'deadbeef', got %x", b)
	}

	tg = NewDbResource(store)
	store.SetPrefix(db.DATATYPE_MENU)
	store.SetLock(db.DATATYPE_MENU, false)
	err = store.Put(ctx, []byte("inky"), []byte("pinky"))
	if err != nil {
		t.Fatal(err)
	}
	store.SetLock(db.DATATYPE_MENU, true)
	rs.WithMenuGetter(tg.GetMenu)

}

func TestDbGetterDirect(t *testing.T) {
	ctx := context.Background()
	store := mem.NewMemDb()
	store.Connect(ctx, "")
	tg := NewDbResource(store)

	store.SetLock(db.DATATYPE_MENU, false)
	store.SetPrefix(db.DATATYPE_MENU)
	err := store.Put(ctx, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	store.SetLock(db.DATATYPE_MENU, true)
	v, err := tg.GetMenu(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if v != "foo" {
		t.Fatalf("expected 'foo', got '%s'", v)
	}

	tg.With(db.DATATYPE_STATICLOAD)
	store.SetLock(db.DATATYPE_STATICLOAD, false)
	store.SetPrefix(db.DATATYPE_STATICLOAD)
	err = store.Put(ctx, []byte("inky.txt"), []byte("blinky"))
	if err != nil {
		t.Fatal(err)
	}
	store.SetLock(db.DATATYPE_STATICLOAD, true)
	fn, err := tg.DbFuncFor(ctx, "inky")
	if err != nil {
		t.Fatal(err)
	}
	r, err := fn(ctx, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Content != "blinky" {
		t.Fatalf("expected 'foo', got '%s'", v)
	}
}
