package mem

import (
	"bytes"
	"context"
	"testing"

	"git.defalsify.org/vise.git/db"
)

func TestCasesMem(t *testing.T) {
	ctx := context.Background()

	store := NewMemDb()
	err := store.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	err = db.RunTests(t, ctx, store)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetMem(t *testing.T) {
	var dbi db.Db
	ctx := context.Background()
	sid := "ses"
	store := NewMemDb()
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sid)

	dbi = store
	_ = dbi

	err := store.Connect(ctx, "")
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
