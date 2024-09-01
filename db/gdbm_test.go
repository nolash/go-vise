package db

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
)

func TestCasesGdbm(t *testing.T) {
	ctx := context.Background()

	db := NewGdbmDb()
	f, err := ioutil.TempFile("", "vise-db-gdbm-*")
	if err != nil {
		t.Fatal(err)
	}
	err = db.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = runTests(t, ctx, db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetGdbm(t *testing.T) {
	var dbi Db
	ctx := context.Background()
	sid := "ses"
	f, err := ioutil.TempFile("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	db := NewGdbmDb()
	db.SetPrefix(DATATYPE_USERDATA)
	db.SetSession(sid)

	dbi = db
	_ = dbi

	err = db.Connect(ctx, f.Name())
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
