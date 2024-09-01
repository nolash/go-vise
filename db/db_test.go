package db

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strconv"
	"testing"

	"git.defalsify.org/vise.git/lang"
)

type testCase struct {
	typ uint8
	s string
	k []byte
	v []byte
	x []byte
	l *lang.Language
}

type testVector struct {
	c map[string]*testCase
	v []string
	i int
	s string
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

func(tc *testCase) Expect() []byte {
	return tc.x
}

func(tv *testVector) add(typ uint8, k string, v string, session string, expect string, language string)  {
	var b []byte
	var x []byte
	var err error
	var ln *lang.Language

	if typ == DATATYPE_BIN {
		b, err = hex.DecodeString(v)
		if err != nil {
			panic(err)
		}
		x, err = hex.DecodeString(expect)
		if err != nil {
			panic(err)
		}
	} else {
		b = []byte(v)
		x = []byte(expect)
	}
	
	if language != "" {
		lo, err := lang.LanguageFromCode(language)
		if err != nil {
			panic(err)
		}
		ln = &lo
	}

	o := &testCase {
		typ: typ,
		k: []byte(k),
		v: b,
		s: session,
		x: x,
		l: ln,
	}
	s := path.Join(strconv.Itoa(int(typ)), session)
	s = path.Join(s, k)
	tv.c[s] = o
	i := len(tv.v)
	tv.v = append(tv.v, s)
	logg.Tracef("add testcase", "i", i, "s", s, "k", o.k)
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

func(tv *testVector) label() string {
	return tv.s
}

func generateSessionTestVectors() testVector {
	tv := testVector{c: make(map[string]*testCase), s: "session"}
	tv.add(DATATYPE_BIN, "foo", "deadbeef", "", "beeffeed", "")
	tv.add(DATATYPE_BIN, "foo", "beeffeed", "inky", "beeffeed", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "tinkywinky", "", "dipsy", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "dipsy", "pinky", "dipsy", "")
	tv.add(DATATYPE_MENU, "foo", "lala", "", "pu", "")
	tv.add(DATATYPE_MENU, "foo", "pu", "blinky", "pu", "")
	tv.add(DATATYPE_STATICLOAD, "foo", "bar", "", "baz", "")
	tv.add(DATATYPE_STATICLOAD, "foo", "baz", "clyde", "baz", "")
	tv.add(DATATYPE_STATE, "foo", "xyzzy", "", "xyzzy", "")
	tv.add(DATATYPE_STATE, "foo", "plugh", "sue", "plugh", "")
	tv.add(DATATYPE_USERDATA, "foo", "itchy", "", "itchy", "")
	tv.add(DATATYPE_USERDATA, "foo", "scratchy", "poochie", "scratchy", "")
	return tv
}

func generateLanguageTestVectors() testVector {
	tv := testVector{c: make(map[string]*testCase), s: "language"}
	tv.add(DATATYPE_BIN, "foo", "deadbeef", "", "beeffeed", "")
	tv.add(DATATYPE_BIN, "foo", "deadbeef", "", "deadbeef", "nor")
	return tv
}

func runTest(t *testing.T, ctx context.Context, db Db, vs testVector) error {
	err := vs.put(ctx, db)
	if err != nil {
		return err
	}
	for true {
		i, tc := vs.next()
		if i == -1 {
			break
		}
		s := fmt.Sprintf("Test%sTyp%dKey%s", vs.label(), tc.Typ(), tc.Key())
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
			if !bytes.Equal(tc.Expect(), v) {
				t.Fatalf("expected %s, got %s", tc.Expect(), v)
			}
		})
		if !r {
			return errors.New("subtest fail")
		}
	}
	return nil

}
func runTests(t *testing.T, ctx context.Context, db Db) error {
	err := runTest(t, ctx, db, generateSessionTestVectors())
	if err != nil {
		return err
	}

	err = runTest(t, ctx, db, generateLanguageTestVectors())
	if err != nil {
		return err
	}
	
	return nil
}
