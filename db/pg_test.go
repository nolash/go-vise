package db

import (
	"bytes"
	"context"
	"testing"
)

func TestCreate(t *testing.T) {
//	t.Skip("need postgresql mock")
	db := NewPgDb().WithSchema("vvise")
	ctx := context.Background()
	err := db.Connect(ctx, "postgres://vise:esiv@localhost:5432/visedb")
	if err != nil {
		t.Fatal(err)
	}
	err = db.Put(ctx, "xyzzy", []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := db.Get(ctx, "xyzzy", []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("bar")) {
		t.Fatalf("expected 'bar', got %x", b)
	}
	err = db.Put(ctx, "xyzzy", []byte("foo"), []byte("plugh"))
	if err != nil {
		t.Fatal(err)
	}
	b, err = db.Get(ctx, "xyzzy", []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, []byte("plugh")) {
		t.Fatalf("expected 'plugh', got %x", b)
	}

}
