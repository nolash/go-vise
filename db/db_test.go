package db

import (
	"context"
	"encoding/hex"
	"path"
)

type testCase struct {
	typ uint8
	s string
	k []byte
	v []byte
}

type testVector struct {
	c map[string]*testCase
	v []string
	i int
}

func(tc *testCase) Key() []byte {
	return tc.k
}

func(tc *testCase) Val() []byte {
	return tc.v
}

func(tc *testCase) Typ() uint8 {
	return tc.typ
}

func(tc *testCase) Session() string {
	return tc.s
}

func(tv *testVector) add(typ uint8, k string, v string, session string) {
	var b []byte
	var err error

	if typ == DATATYPE_BIN {
		b, err = hex.DecodeString(v)
		if err != nil {
			panic(err)
		}
	} else {
		b = []byte(v)
	}

	o := &testCase {
		typ: typ,
		k: []byte(k),
		v: b,
		s: session,
	}
	s := path.Join(session, k)
	tv.c[s] = o
	tv.v = append(tv.v, s)
}

func(tv *testVector) next() (int, *testCase) {
	i := tv.i
	if i == len(tv.v) {
		return -1, nil
	}
	tv.i++
	return i, tv.c[tv.v[i]]
}

func(tv *testVector) rewind() {
	tv.i = 0
}

func(tv *testVector) put(ctx context.Context, db Db) error {
	var i int
	var tc *testCase
	defer tv.rewind()

	for true {
		i, tc = tv.next()
		if i == -1 {
			break
		}
		db.SetPrefix(tc.Typ())
		db.SetSession(tc.Session())
		db.SetLock(tc.Typ(), false)
		err := db.Put(ctx, tc.Key(), tc.Val())
		if err != nil {
			return err
		}
		db.SetLock(tc.Typ(), true)
	}
	return nil
}

func generateTestVectors() testVector {
	tv := testVector{c: make(map[string]*testCase)}
	tv.add(DATATYPE_BIN, "foo", "deadbeef", "")
	tv.add(DATATYPE_BIN, "foo", "beeffeed", "tinkywinky")
	tv.add(DATATYPE_TEMPLATE, "foo", "inky", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "pinky", "dipsy")
	tv.add(DATATYPE_MENU, "foo", "blinky", "")
	tv.add(DATATYPE_MENU, "foo", "clyde", "lala")
	tv.add(DATATYPE_STATICLOAD, "foo", "bar", "")
	tv.add(DATATYPE_STATICLOAD, "foo", "baz", "po")
	return tv
}

