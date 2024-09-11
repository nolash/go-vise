package fs

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/db/dbtest"
)

func TestCasesFs(t *testing.T) {
	ctx := context.Background()

	store := NewFsDb()
	d, err := ioutil.TempDir("", "vise-db-fs-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}

	err = dbtest.RunTests(t, ctx, store)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetFs(t *testing.T) {
	var dbi db.Db
	ctx := context.Background()
	sid := "ses"
	d, err := ioutil.TempDir("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	store := NewFsDb()
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sid)

	dbi = store
	_ = dbi

	err = store.Connect(ctx, d)
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

func TestPutGetFsAlt(t *testing.T) {
	ctx := context.Background()
	sid := "zezion"
	d, err := ioutil.TempDir("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	store := NewFsDb()
	store.SetPrefix(db.DATATYPE_TEMPLATE)
	store.SetSession(sid)

	fp := path.Join(d, sid)
	err = os.MkdirAll(fp, 0700)
	if err != nil {
		t.Fatal(err)
	}
	store.Connect(ctx, fp)
	fp = path.Join(fp, "inky")

	b := []byte("pinky blinky clyde")
	err = ioutil.WriteFile(fp, b, 0700)
	if err != nil {
		t.Fatal(err)
	}
	
	v, err := store.Get(ctx, []byte("inky"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, b) {
		t.Fatalf("expected %x, got %x", b, v)
	}
}

func TestConnect(t *testing.T) {
	ctx := context.Background()
	store := NewFsDb()
	err := store.Connect(ctx, "")
	if err == nil {
		t.Fatal("expected error")
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
}

func TestReopen(t *testing.T) {
	ctx := context.Background()
	store := NewFsDb()
	d, err := ioutil.TempDir("", "vise-db-fs-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, d)
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

	store = NewFsDb()
	err = store.Connect(ctx, d)
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
	store := NewFsDb()
	d, err := ioutil.TempDir("", "vise-db-fs-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Get(ctx, []byte("xyzzy"))
	if err == nil {
		t.Fatal(err)
	}
}
