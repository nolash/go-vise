package gdbm

import (
	"bytes"
	"context"
	"errors"

	gdbm "github.com/graygnuorg/go-gdbm"
	
	"git.defalsify.org/vise.git/db"
)

func(gdb *gdbmDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	gdb.SetLanguage(nil)
	lk, err := gdb.ToKey(ctx, key)
	if err != nil {
		return nil, err
	}
	key = lk.Default
	
	gdb.it = gdb.conn.Iterator()
	for true {
		k, err := gdb.it()
		if err != nil {
			if errors.Is(err, gdbm.ErrItemNotFound) {
				err = db.NewErrNotFound(key)
			}
			gdb.it = nil
			return nil, err
		}
		if !bytes.HasPrefix(k, key) {
			continue
		}
		kk, err := gdb.DecodeKey(ctx, k)
		if err != nil {
			return nil, err
		}
		v, err := gdb.Get(ctx, kk)
		if err != nil {
			gdb.it = nil
			return nil, err
		}
		gdb.itBase = key
		return db.NewDumper(gdb.dumpFunc).WithFirst(kk, v), nil
	}
	gdb.it = nil
	return nil, db.NewErrNotFound(key)
}

func(gdb *gdbmDb) dumpFunc(ctx context.Context) ([]byte, []byte) {
	var k []byte
	var match bool
	var err error

	for true {
		k, err = gdb.it()
		if err != nil {
			gdb.it = nil
			return nil, nil
		}
		if bytes.HasPrefix(k, gdb.itBase) {
			match = true
			break
		}
	}
	if !match {
		gdb.it = nil
		return nil, nil
	}
	kk, err := gdb.DecodeKey(ctx, k)
	if err != nil {
		return nil, nil
	}
	v, err := gdb.Get(ctx, kk)
	if err != nil {
		return nil, nil
	}
	return kk, v
}
