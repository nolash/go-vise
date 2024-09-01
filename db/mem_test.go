package db

import (
	"bytes"
	"context"
	"testing"
)

func TestCasesMem(t *testing.T) {
	ctx := context.Background()

	db := NewMemDb()
	err := db.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	err = runTests(t, ctx, db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetMem(t *testing.T) {
	var dbi Db
	ctx := context.Background()
	sid := "ses"
	db := NewMemDb()
	db.SetPrefix(DATATYPE_USERDATA)
	db.SetSession(sid)

	dbi = db
	_ = dbi

	err := db.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	err = db.Put(ctx, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := db.Get(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("bar")) {
		t.Fatalf("expected value 'bar', found '%s'", v)
	}
	_, err = db.Get(ctx, []byte("bar"))
	if err == nil {
		t.Fatal("expected get error for key 'bar'")
	}
}
