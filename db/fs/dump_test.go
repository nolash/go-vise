package fs

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"git.defalsify.org/vise.git/db"
)

func TestDumpFs(t *testing.T) {
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
	if !bytes.Equal(k, []byte("foobar")) {
		t.Fatalf("expected key 'foobar', got %s", k)
	}
	if !bytes.Equal(v, []byte("pinky")) {
		t.Fatalf("expected val 'pinky', got %s", v)
	}
	k, v = o.Next(ctx)
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

func TestDumpBinary(t *testing.T) {
	ctx := context.Background()

	store := NewFsDb()
	store = store.WithBinary()
	d, err := ioutil.TempDir("", "vise-db-fsbin-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}

	store.SetPrefix(db.DATATYPE_USERDATA)
	err = store.Put(ctx, []byte{0x01, 0x02, 0x03}, []byte("inky"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte{0x01, 0x02, 0x04}, []byte("pinky"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte{0x02, 0x03, 0x04}, []byte("blinky"))
	if err != nil {
		t.Fatal(err)
	}
	o, err := store.Dump(ctx, []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	k, v := o.Next(ctx)
	if !bytes.Equal(k, []byte{0x01, 0x02, 0x03}) {
		t.Fatalf("expected key '0x010203', got %x", k)
	}
	if !bytes.Equal(v, []byte("inky")) {
		t.Fatalf("expected val 'inky', got %s", v)
	}
	k, v = o.Next(ctx)
	if !bytes.Equal(k, []byte{0x01, 0x02, 0x04}) {
		t.Fatalf("expected key '0x010204', got %x", k)
	}
	if !bytes.Equal(v, []byte("pinky")) {
		t.Fatalf("expected val 'pinky', got %s", v)
	}
	k, v = o.Next(ctx)
	if k != nil {
		t.Fatalf("expected nil, got %s", k)
	}
}

func TestDumpSessionBinary(t *testing.T) {
	ctx := context.Background()

	store := NewFsDb()
	store = store.WithBinary()
	store.SetSession("foobar")
	d, err := ioutil.TempDir("", "vise-db-fsbin-*")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}

	store.SetPrefix(db.DATATYPE_USERDATA)
	err = store.Put(ctx, []byte{0x01, 0x02, 0x03}, []byte("inky"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte{0x01, 0x02, 0x04}, []byte("pinky"))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, []byte{0x02, 0x03, 0x04}, []byte("blinky"))
	if err != nil {
		t.Fatal(err)
	}
	o, err := store.Dump(ctx, []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	k, v := o.Next(ctx)
	if !bytes.Equal(k, []byte{0x01, 0x02, 0x03}) {
		t.Fatalf("expected key '0x010203', got %x", k)
	}
	if !bytes.Equal(v, []byte("inky")) {
		t.Fatalf("expected val 'inky', got %s", v)
	}
	k, v = o.Next(ctx)
	if !bytes.Equal(k, []byte{0x01, 0x02, 0x04}) {
		t.Fatalf("expected key '0x010204', got %x", k)
	}
	if !bytes.Equal(v, []byte("pinky")) {
		t.Fatalf("expected val 'pinky', got %s", v)
	}
	k, v = o.Next(ctx)
	if k != nil {
		t.Fatalf("expected nil, got %s", k)
	}
}
