package db

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

func TestCasesMem(t *testing.T) {
	var i int
	var tc *testCase

	ctx := context.Background()
	vs := generateTestVectors()
	db := NewMemDb()
	err := db.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	err = vs.put(ctx, db)

	for true {
		i, tc = vs.next()
		if i == -1 {
			break
		}
		s := fmt.Sprintf("TestTyp%dKey%s", tc.Typ(), tc.Key())
		if tc.Session() != "" {
			s += "Session" + tc.Session()
		} else {
			s += "NoSession"
		}
		r := t.Run(s, func(t *testing.T) {
			db.SetPrefix(tc.Typ())
			db.SetSession(tc.Session())
			v, err := db.Get(ctx, tc.Key())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(tc.Val(), v) {
				t.Fatalf("expected %x, got %x", tc.Val(), v)
			}
		})
		if !r {
			t.Fatalf("subtest fail")
		}
	}
}

func TestPutGetMem(t *testing.T) {
	var dbi Db
	ctx := context.Background()
	sid := "ses"
	db := NewMemDb()
	db.SetPrefix(DATATYPE_USERSTART)
	db.SetSession(sid)

	dbi = db
	_ = dbi

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
