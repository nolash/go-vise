package db

import (
	"bytes"
	"context"
	"testing"
)

func TestCasesPg(t *testing.T) {
	ctx := context.Background()

	db := NewPgDb().WithSchema("vvise")
	t.Skip("need postgresql mock")

	err := db.Connect(ctx, "postgres://vise:esiv@localhost:5432/visedb")
	if err != nil {
		t.Fatal(err)
	}

	err = runTests(t, ctx, db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetPg(t *testing.T) {
	var dbi Db
	ses := "xyzzy"
	db := NewPgDb().WithSchema("vvise")
	db.SetPrefix(DATATYPE_USERDATA)
	db.SetSession(ses)
	ctx := context.Background()

	dbi = db
	_ = dbi

	t.Skip("need postgresql mock")
	err := db.Connect(ctx, "postgres://vise:esiv@localhost:5432/visedb")
	if err != nil {
		t.Fatal(err)
	}
	err = db.Put(ctx, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := db.Get(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("bar")) {
		t.Fatalf("expected 'bar', got %x", b)
	}
	err = db.Put(ctx, []byte("foo"), []byte("plugh"))
	if err != nil {
		t.Fatal(err)
	}
	b, err = db.Get(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("plugh")) {
		t.Fatalf("expected 'plugh', got %x", b)
	}

}
