package log

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/db/mem"
)

func TestLogDb(t *testing.T) {
	sessionId := "xyzzy"
	ctx := context.Background()
	main := mem.NewMemDb()
	sub := mem.NewMemDb()
	store := NewLogDb(main, sub)
	err := store.Connect(ctx, "main")
	if err != nil {
		t.Fatal(err)
	}

	k := []byte("foo")
	v := []byte("bar")
	tstart := time.Now()
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}

	r, err := store.Get(ctx, k)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(r, v) {
		t.Fatalf("Expected %x, got %x", v, r)
	}

	sub.SetPrefix(db.DATATYPE_UNKNOWN)
	tend := time.Now()
	dump, err := sub.Dump(ctx, append([]byte{db.DATATYPE_UNKNOWN}, []byte(sessionId)...))
	if err != nil {
		t.Fatal(err)
	}
	r, _ = dump.Next(ctx)
	targetLen := len(sessionId) + 8 + 1 + 1
	if len(r) != targetLen {
		t.Fatalf("Unexpected length %d (%x), should be %d", len(r), r, targetLen)
	}

	k, err = sub.FromSessionKey(r[1:])
	if err != nil {
		t.Fatal(err)
	}
	tn := binary.BigEndian.Uint64(k)
	tExpect := uint64(tstart.UnixNano())
	if tn <= tExpect {
		t.Fatalf("expected %d should be after %d", tn, tExpect)
	}
	tExpect = uint64(tend.UnixNano())
	if tn >= tExpect {
		t.Fatalf("expected %v should be before %v", tn, tExpect)
	}

}
