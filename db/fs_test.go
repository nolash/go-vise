package db

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestCasesFs(t *testing.T) {
	ctx := context.Background()

	db := NewFsDb()
	d, err := ioutil.TempDir("", "vise-db-fs-*")
	if err != nil {
		t.Fatal(err)
	}
	err = db.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}

	err = runTests(t, ctx, db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetFs(t *testing.T) {
	var dbi Db
	ctx := context.Background()
	sid := "ses"
	d, err := ioutil.TempDir("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	db := NewFsDb()
	db.SetPrefix(DATATYPE_USERDATA)
	db.SetSession(sid)

	dbi = db
	_ = dbi

	err = db.Connect(ctx, d)
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

func TestPutGetFsAlt(t *testing.T) {
	ctx := context.Background()
	sid := "zezion"
	d, err := ioutil.TempDir("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	db := NewFsDb()
	db.SetPrefix(DATATYPE_TEMPLATE)
	db.SetSession(sid)

	fp := path.Join(d, sid)
	err = os.MkdirAll(fp, 0700)
	if err != nil {
		t.Fatal(err)
	}
	db.Connect(ctx, fp)
	fp = path.Join(fp, "inky")

	b := []byte("pinky blinky clyde")
	err = ioutil.WriteFile(fp, b, 0700)
	if err != nil {
		t.Fatal(err)
	}
	
	v, err := db.Get(ctx, []byte("inky"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, b) {
		t.Fatalf("expected %x, got %x", b, v)
	}
}
