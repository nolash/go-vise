package db

import (
	"bytes"
	"context"
	"testing"
)

func TestPutGetMem(t *testing.T) {
	ctx := context.Background()
	sid := "ses"
	db := &MemDb{}
	db.SetPrefix(DATATYPE_USERSTART)
	db.SetSession(sid)
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
