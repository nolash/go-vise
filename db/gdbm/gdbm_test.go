package gdbm

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/db/dbtest"
)

func TestCasesGdbm(t *testing.T) {
	ctx := context.Background()

	store := NewGdbmDb()
	f, err := ioutil.TempFile("", "vise-db-gdbm-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = dbtest.RunTests(t, ctx, store)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetGdbm(t *testing.T) {
	var dbi db.Db
	ctx := context.Background()
	sid := "ses"
	f, err := ioutil.TempFile("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	store := NewGdbmDb()
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sid)

	dbi = store
	_ = dbi

	err = store.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := store.Get(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("bar")) {
		t.Fatalf("expected value 'bar', found '%s'", v)
	}
	_, err = store.Get(ctx, []byte("bar"))
	if err == nil {
		t.Fatal("expected get error for key 'bar'")
	}

}

func TestConnect(t *testing.T) {
	ctx := context.Background()
	store := NewGdbmDb()
	err := store.Connect(ctx, "")
	if err == nil {
		t.Fatal("expected error")
	}
	f, err := ioutil.TempFile("", "vise-db-gdbm-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}
	err = store.SetLock(db.DATATYPE_USERDATA, true)
	if err != nil {
		t.Fatal(err)
	}
	store.SetPrefix(db.DATATYPE_USERDATA)
	if store.CheckPut() {
		t.Fatal("expected checkput false")
	}
	err = store.Put(ctx, []byte("foo"), []byte("bar"))
	if err == nil {
		t.Fatal("expected error")
	}
	if store.CheckPut() {
		t.Fatal("expected checkput false")
	}
	err = store.SetLock(db.DATATYPE_USERDATA, false)
	if err != nil {
		t.Fatal(err)
	}
	if !store.CheckPut() {
		t.Fatal("expected checkput false")
	}
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte("foo"), []byte("bar"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReopen(t *testing.T) {
	ctx := context.Background()
	store := NewGdbmDb()
	f, err := ioutil.TempFile("", "vise-db-gdbm-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}
	store.SetPrefix(db.DATATYPE_USERDATA)
	err = store.Put(ctx, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}

	store = NewGdbmDb()
	err = store.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}
	store.SetPrefix(db.DATATYPE_USERDATA)
	v, err := store.Get(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("bar")) {
		t.Fatalf("expected 'bar', got: '%s'", v)
	}
}

func TestNoKey(t *testing.T) {
	ctx := context.Background()
	store := NewGdbmDb()
	f, err := ioutil.TempFile("", "vise-db-gdbm-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, f.Name())
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Get(ctx, []byte("xyzzy"))
	if err == nil {
		t.Fatal(err)
	}
}
