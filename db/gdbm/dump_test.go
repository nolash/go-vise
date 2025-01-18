package gdbm

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"git.defalsify.org/vise.git/db"
)

func TestDumpGdbm(t *testing.T) {
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

	store.SetPrefix(db.DATATYPE_USERDATA)
	err = store.Put(ctx, []byte("bar"), []byte("inky"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte("foobar"), []byte("pinky"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte("foobarbaz"), []byte("blinky"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte("xyzzy"), []byte("clyde"))
	if err != nil {
		t.Fatal(err)
	}

	o, err := store.Dump(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	k, v := o.Next(ctx)
	//if !bytes.Equal(k, append([]byte{db.DATATYPE_USERDATA}, []byte("foobar")...)) {
	if !bytes.Equal(k, []byte("foobar")) {
		t.Fatalf("expected key 'foobar', got '%s'", k)
	}
	if !bytes.Equal(v, []byte("pinky")) {
		t.Fatalf("expected val 'pinky', got %s", v)
	}
	k, v = o.Next(ctx)
	//if !bytes.Equal(k, append([]byte{db.DATATYPE_USERDATA}, []byte("foobarbaz")...)) {
	if !bytes.Equal(k, []byte("foobarbaz")) {
		t.Fatalf("expected key 'foobarbaz', got %s", k)
	}
	if !bytes.Equal(v, []byte("blinky")) {
		t.Fatalf("expected val 'blinky', got %s", v)
	}
	k, v = o.Next(ctx)
	if k != nil {
		t.Fatalf("expected nil, got %s", k)
	}
}
