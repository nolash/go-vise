package db

import (
	"context"
	"encoding/hex"
	"errors"
)

// memDb is a memory backend implementation of the Db interface.
type memDb struct {
	baseDb
	store map[string][]byte
}

// NewmemDb returns an in-process volatile Db implementation.
func NewMemDb(ctx context.Context) *memDb {
	db := &memDb{}
	db.baseDb.defaultLock()
	return db
}

// Connect implements Db
func(mdb *memDb) Connect(ctx context.Context, connStr string) error {
	if mdb.store != nil {
		panic("already connected")
	}
	mdb.store = make(map[string][]byte)
	return nil
}

// convert to a supported map key type
func(mdb *memDb) toHexKey(key []byte) (string, error) {
	k, err := mdb.ToKey(key)
	return hex.EncodeToString(k), err
}

// Get implements Db
func(mdb *memDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	k, err := mdb.toHexKey(key)
	if err != nil {
		return nil, err
	}
	logg.TraceCtxf(ctx, "mem get", "k", k)
	v, ok := mdb.store[k]
	if !ok {
		b, _ := hex.DecodeString(k)
		return nil, NewErrNotFound(b)
	}
	return v, nil
}

// Put implements Db
func(mdb *memDb) Put(ctx context.Context, key []byte, val []byte) error {
	if !mdb.checkPut() {
		return errors.New("unsafe put and safety set")
	}
	k, err := mdb.toHexKey(key)
	if err != nil {
		return err
	}
	mdb.store[k] = val
	logg.TraceCtxf(ctx, "mem put", "k",  k, "v", val)
	return nil
}

// Close implements Db
func(mdb *memDb) Close() error {
	return nil
}
