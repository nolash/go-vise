package db

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
)

func TestPutGetFs(t *testing.T) {
	var dbi Db
	ctx := context.Background()
	sid := "ses"
	d, err := ioutil.TempDir("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	db := NewFsDb()
	db.SetPrefix(DATATYPE_USERSTART)
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
