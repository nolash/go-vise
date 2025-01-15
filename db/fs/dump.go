package fs

import (
	"bytes"
	"context"
	"os"

	"git.defalsify.org/vise.git/db"
)

func(fdb *fsDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	var err error
	key = append([]byte{db.DATATYPE_USERDATA}, key...)
	fdb.matchPrefix = key
	fdb.elements, err = os.ReadDir(fdb.dir)
	if err != nil {
		return nil, err
	}

	if len(fdb.elements) > 0 {
		if len(key) == 0 {
			v := fdb.elements[0]
			fdb.elements = fdb.elements[1:]
			s := v.Name()
			k := []byte(s)
			k[0] -= 0x30
			kk, err := fdb.DecodeKey(ctx, k)
			if err != nil {
				return nil, err
			}
			vv, err := fdb.Get(ctx, kk)
			if err != nil {
				return nil, err
			}
			return db.NewDumper(fdb.dumpFunc).WithFirst(kk, vv), nil
		}
	}
	for len(fdb.elements) > 0 {
		v := fdb.elements[0]
		fdb.elements = fdb.elements[1:]
		s := v.Name()
		k := []byte(s)
		if len(key) > len(k) {
			continue
		}
		k[0] -= 0x30
		kk, err := fdb.DecodeKey(ctx, k)
		if err != nil {
			continue
		}
		kkk := append([]byte{k[0]}, kk...)
		if bytes.HasPrefix(kkk, key) {
			vv, err := fdb.Get(ctx, kk)
			if err != nil {
				return nil, err
			}
			return db.NewDumper(fdb.dumpFunc).WithFirst(kk, vv), nil
		}
	}
	return nil, db.NewErrNotFound(key)
}

func(fdb *fsDb) dumpFunc(ctx context.Context) ([]byte, []byte) {
	if len(fdb.elements) == 0 {
		return nil, nil
	}
	v := fdb.elements[0]
	fdb.elements = fdb.elements[1:]
	s := v.Name()
	k := []byte(s)
	k[0] -= 0x30
	kk, err := fdb.DecodeKey(ctx, k)
	if err != nil {
		return nil, nil
	}
	kkk := append([]byte{k[0]}, kk...)
	if bytes.HasPrefix(kkk, fdb.matchPrefix) {
		vv, err := fdb.Get(ctx, kk)
		if err != nil {
			return nil, nil
		}
		return kk, vv
	}
	return nil, nil
}
