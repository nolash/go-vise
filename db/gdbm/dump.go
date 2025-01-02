package gdbm

import (
	"bytes"
	"context"
	"errors"

	gdbm "github.com/graygnuorg/go-gdbm"
	
	"git.defalsify.org/vise.git/db"
)

func(gdb *gdbmDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	key = append([]byte{db.DATATYPE_USERDATA}, key...)
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
		logg.TraceCtxf(ctx, "dump trace", "k", k, "key", key)
		if !bytes.HasPrefix(k, key) {
			continue
		}
		gdb.SetPrefix(k[0])
		v, err := gdb.Get(ctx, k[1:])
		if err != nil {
			gdb.it = nil
			return nil, err
		}
		gdb.itBase = key
		return db.NewDumper(gdb.dumpFunc).WithFirst(k, v), nil
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
		//if bytes.HasPrefix(k[1:], gdb.itBase) {
		if bytes.HasPrefix(k, gdb.itBase) {
			match = true
			break
		}
	}
	if !match {
		gdb.it = nil
		return nil, nil
	}
	v, err := gdb.Get(ctx, k[1:])
	if err != nil {
		return nil, nil
	}
	return k, v
}

//func(gdb *gdbmDb) After(ctx context.Context, keyPart []byte) ([]byte, []byte) {
//	if keyPart == nil {
//		gdb.it = gdb.conn.Iterator()
//		return nil, nil
//	}
//	k, err := gdb.it()
//	if err != nil {
//		if !errors.Is(err, gdbm.ErrItemNotFound) {
//			panic(err)
//		}
//		gdb.it = gdb.conn.Iterator()
//	}
//	v, err := gdb.Get(ctx, k)
//	if err != nil {
//		panic(err)
//	}
//	return k, v
//}
