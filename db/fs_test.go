package db

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
)

func TestPutGet(t *testing.T) {
	ctx := context.Background()
	sid := "ses"
	d, err := ioutil.TempDir("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	db := &FsDb{}
	err = db.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Put(ctx, sid, []byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := db.Get(ctx, sid, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("bar")) {
		t.Fatalf("expected value 'bar', found '%s'", v)
	}
	_, err = db.Get(ctx, sid, []byte("bar"))
	if err == nil {
		t.Fatal("expected get error for key 'bar'")
	}
}
