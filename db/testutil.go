package db

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
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
	t string
}

type testVector struct {
	c map[string]*testCase
	v []string
	i int
	s string
}

type testFunc func() testVector

var (
	tests = []testFunc{
		generateSessionTestVectors,
		generateMultiSessionTestVectors,
		generateLanguageTestVectors,
		generateMultiLanguageTestVectors,
		generateSessionLanguageTestVectors,
	}
	dataTypeDebug = map[uint8]string{
		DATATYPE_BIN: "bytecode",
		DATATYPE_TEMPLATE: "template",
		DATATYPE_MENU: "menu",
		DATATYPE_STATICLOAD: "staticload",
		DATATYPE_STATE: "state",
		DATATYPE_USERDATA: "udata",
	}
)

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

func(tc *testCase) Lang() string {
	if tc.l == nil {
		return ""
	}
	return tc.l.Code
}

func(tc *testCase) Expect() []byte {
	return tc.x
}

func(tc *testCase) Label() string {
	return tc.t
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
	s := dataTypeDebug[typ]
	s = path.Join(s, session)
	s = path.Join(s, k)
	if ln != nil {
		s = path.Join(s, language)
	}
	o := &testCase {
		typ: typ,
		k: []byte(k),
		v: b,
		s: session,
		x: x,
		l: ln,
		t: s,
	}
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
		logg.TraceCtxf(ctx, "running put for test", "vector", tv.label(), "case", tc.Label())
		db.SetPrefix(tc.Typ())
		db.SetSession(tc.Session())
		db.SetLock(tc.Typ(), false)
		db.SetLanguage(nil)
		if tc.Lang() != "" {
			ln, err := lang.LanguageFromCode(tc.Lang())
			if err != nil {
				return err
			}
			db.SetLanguage(&ln)
		}
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
	tv.add(DATATYPE_BIN, "foo", "beeffeed", "", "beeffeed", "nor")
	tv.add(DATATYPE_TEMPLATE, "foo", "tinkywinky", "", "tinkywinky", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "dipsy", "", "dipsy", "nor")
	tv.add(DATATYPE_MENU, "foo", "lala", "", "lala", "")
	tv.add(DATATYPE_MENU, "foo", "pu", "", "pu", "nor")
	tv.add(DATATYPE_STATICLOAD, "foo", "bar", "", "bar", "")
	tv.add(DATATYPE_STATICLOAD, "foo", "baz", "", "baz", "nor")
	tv.add(DATATYPE_STATE, "foo", "xyzzy", "", "plugh", "")
	tv.add(DATATYPE_STATE, "foo", "plugh", "", "plugh", "nor")
	tv.add(DATATYPE_USERDATA, "foo", "itchy", "", "scratchy", "")
	tv.add(DATATYPE_USERDATA, "foo", "scratchy", "", "scratchy", "nor")
	return tv
}

func generateMultiLanguageTestVectors() testVector {
	tv := testVector{c: make(map[string]*testCase), s: "multilanguage"}
	tv.add(DATATYPE_TEMPLATE, "foo", "tinkywinky", "", "pu", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "dipsy", "", "dipsy", "nor")
	tv.add(DATATYPE_TEMPLATE, "foo", "lala", "", "lala", "swa")
	tv.add(DATATYPE_TEMPLATE, "foo", "pu", "", "pu", "")
	return tv
}

func generateSessionLanguageTestVectors() testVector {
	tv := testVector{c: make(map[string]*testCase), s: "sessionlanguage"}
	tv.add(DATATYPE_TEMPLATE, "foo", "tinkywinky", "", "pu", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "dipsy", "", "lala", "nor")
	tv.add(DATATYPE_TEMPLATE, "foo", "lala", "bar", "lala", "nor")
	tv.add(DATATYPE_TEMPLATE, "foo", "pu", "bar", "pu", "")
	tv.add(DATATYPE_STATE, "foo", "inky", "", "pinky", "")
	tv.add(DATATYPE_STATE, "foo", "pinky", "", "pinky", "nor")
	tv.add(DATATYPE_STATE, "foo", "blinky", "bar", "clyde", "nor")
	tv.add(DATATYPE_STATE, "foo", "clyde", "bar", "clyde", "")
	tv.add(DATATYPE_BIN, "foo", "deadbeef", "", "feebdaed", "")
	tv.add(DATATYPE_BIN, "foo", "beeffeed", "", "feebdaed", "nor")
	tv.add(DATATYPE_BIN, "foo", "deeffeeb", "baz", "feebdaed", "nor")
	tv.add(DATATYPE_BIN, "foo", "feebdaed", "baz", "feebdaed", "")
	return tv
}

func generateMultiSessionTestVectors() testVector {
	tv := testVector{c: make(map[string]*testCase), s: "multisession"}
	tv.add(DATATYPE_TEMPLATE, "foo", "red", "", "blue", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "green", "bar", "blue", "")
	tv.add(DATATYPE_TEMPLATE, "foo", "blue", "baz", "blue", "")
	tv.add(DATATYPE_STATE, "foo", "inky", "", "inky", "")
	tv.add(DATATYPE_STATE, "foo", "pinky", "clyde", "pinky", "")
	tv.add(DATATYPE_STATE, "foo", "blinky", "sue", "blinky", "")
	tv.add(DATATYPE_BIN, "foo", "deadbeef", "", "feebdeef", "")
	tv.add(DATATYPE_BIN, "foo", "feedbeef", "bar", "feebdeef", "")
	tv.add(DATATYPE_BIN, "foo", "feebdeef", "baz", "feebdeef", "")
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
		s := fmt.Sprintf("Test%s[%d]%s", vs.label(), i, tc.Label())
		r := t.Run(s, func(t *testing.T) {
			db.SetPrefix(tc.Typ())
			db.SetSession(tc.Session())
			if tc.Lang() != "" {
				ln, err := lang.LanguageFromCode(tc.Lang())
				if err != nil {
					t.Fatal(err)
				}
				db.SetLanguage(&ln)
			} else {
				db.SetLanguage(nil)
			}
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
	for _, fn := range tests {
		err := runTest(t, ctx, db, fn())
		if err != nil {
			return err
		}
	}
	
	return nil
}

func RunTests(t *testing.T, ctx context.Context, db Db) error {
	return runTests(t, ctx, db)
}
