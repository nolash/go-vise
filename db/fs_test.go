package db

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestPutGet(t *testing.T) {
	d, err := ioutil.TempDir("", "vise-db-*")
	if err != nil {
		t.Fatal(err)
	}
	db := &FsDb{}
	err = db.Connect(d)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Put([]byte("foo"), []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := db.Get([]byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("bar")) {
		t.Fatalf("expected value 'bar', found '%s'", v)
	}
	_, err = db.Get([]byte("bar"))
	if err == nil {
		t.Fatal("expected get error for key 'bar'")
	}
}
