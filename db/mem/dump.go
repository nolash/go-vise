package mem

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"

	"git.defalsify.org/vise.git/db"
)

// Dump implements Db.
func (mdb *memDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	mdb.dumpKeys = slices.Sorted(maps.Keys(mdb.store))
	mdb.dumpIdx = -1
	for i := 0; i < len(mdb.dumpKeys); i++ {
		s := mdb.dumpKeys[i]
		k, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}
		if bytes.HasPrefix(k, key) {
			logg.DebugCtxf(ctx, "starting dump", "key", k)
			mdb.dumpIdx = i
			kk, err := mdb.Base().FromSessionKey(k[1:])
			if err != nil {
				return nil, fmt.Errorf("invalid dump key %x: %v", k, err)
			}
			v, err := mdb.Get(ctx, kk[:])
			if err != nil {
				return nil, fmt.Errorf("value err for key %x: %v", k, err)
			}
			return db.NewDumper(mdb.dumpFunc).WithFirst(k, v), nil
		}
	}
	return nil, db.NewErrNotFound(key)
}

func (mdb *memDb) dumpFunc(ctx context.Context) ([]byte, []byte) {
	if mdb.dumpIdx == -1 {
		return nil, nil
	}
	if mdb.dumpIdx >= len(mdb.dumpKeys) {
		mdb.dumpIdx = -1
		return nil, nil
	}
	s := mdb.dumpKeys[mdb.dumpIdx]
	k, err := hex.DecodeString(s)
	if err != nil {
		mdb.dumpIdx = -1
		return nil, nil
	}
	kk, err := mdb.Base().FromSessionKey(k[1:])
	if err != nil {
		return nil, nil
	}
	v, err := mdb.Get(ctx, kk)
	if err != nil {
		return nil, nil
	}
	return k, v
}
