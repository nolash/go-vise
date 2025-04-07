package mem

import (
	"context"
	"encoding/hex"
	"errors"

	"git.defalsify.org/vise.git/db"
)

// holds string (hex) versions of lookupKey
type memLookupKey struct {
	Default     string
	Translation string
}

// memDb is a memory backend implementation of the Db interface.
type memDb struct {
	*db.DbBase
	store map[string][]byte
	dumpIdx int
	dumpKeys []string
}

// NewmemDb returns an in-process volatile Db implementation.
func NewMemDb() *memDb {
	db := &memDb{
		DbBase: db.NewDbBase(),
		dumpIdx: -1,
	}
	return db
}

// Base implements Db
func (mdb *memDb) Base() *db.DbBase {
	return mdb.DbBase
}

// String implements the string interface.
func (mdb *memDb) String() string {
	return "memdb"
}

// Connect implements Db
func (mdb *memDb) Connect(ctx context.Context, connStr string) error {
	if mdb.store != nil {
		logg.WarnCtxf(ctx, "already connected")
		return nil
	}
	mdb.store = make(map[string][]byte)
	return nil
}

// convert to a supported map key type
func (mdb *memDb) toHexKey(ctx context.Context, key []byte) (memLookupKey, error) {
	var mk memLookupKey
	lk, err := mdb.ToKey(ctx, key)
	mk.Default = hex.EncodeToString(lk.Default)
	if lk.Translation != nil {
		mk.Translation = hex.EncodeToString(lk.Translation)
	}
	logg.TraceCtxf(ctx, "converted key", "orig", key, "b", lk, "s", mk)
	return mk, err
}

// Get implements Db
func (mdb *memDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	var v []byte
	var ok bool
	mk, err := mdb.toHexKey(ctx, key)
	if err != nil {
		return nil, err
	}
	logg.TraceCtxf(ctx, "mem get", "k", mk)
	if mk.Translation != "" {
		v, ok = mdb.store[mk.Translation]
		if ok {
			return v, nil
		}
	}
	v, ok = mdb.store[mk.Default]
	if !ok {
		//b, _ := hex.DecodeString(k)
		return nil, db.NewErrNotFound(key)
	}
	return v, nil
}

// Put implements Db
func (mdb *memDb) Put(ctx context.Context, key []byte, val []byte) error {
	var k string
	if !mdb.CheckPut() {
		return errors.New("unsafe put and safety set")
	}
	mk, err := mdb.toHexKey(ctx, key)
	if err != nil {
		return err
	}
	if mk.Translation != "" {
		k = mk.Translation
	} else {
		k = mk.Default
	}
	mdb.store[k] = val
	logg.TraceCtxf(ctx, "mem put", "k", k, "mk", mk, "v", val)
	return nil
}

// Close implements Db
func (mdb *memDb) Close(ctx context.Context) error {
	return nil
}
