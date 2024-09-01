package db

import (
	"bytes"
	"context"
	"testing"

	"git.defalsify.org/vise.git/db"
)

func TestCasesPg(t *testing.T) {
	ctx := context.Background()

	store := NewPgDb().WithSchema("vvise")
	t.Skip("need postgresql mock")

	err := store.Connect(ctx, "postgres://vise:esiv@localhost:5432/visedb")
	if err != nil {
		t.Fatal(err)
	}

	err = db.RunTests(t, ctx, store)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetPg(t *testing.T) {
	var dbi db.Db
	ses := "xyzzy"
	store := NewPgDb().WithSchema("vvise")
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(ses)
	ctx := context.Background()

	dbi = store 
	_ = dbi

	t.Skip("need postgresql mock")
	err := store.Connect(ctx, "postgres://vise:esiv@localhost:5432/visedb")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.Get(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("bar")) {
		t.Fatalf("expected 'bar', got %x", b)
	}
	err = store.Put(ctx, []byte("foo"), []byte("plugh"))
	if err != nil {
		t.Fatal(err)
	}
	b, err = store.Get(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("plugh")) {
		t.Fatalf("expected 'plugh', got %x", b)
	}

}
