package db

import (
	"context"
	"encoding/hex"
	"errors"
)

// MemDb is a memory backend implementation of the Db interface.
type MemDb struct {
	BaseDb
	store map[string][]byte
}

// NewMemDb returns an already allocated 
func NewMemDb(ctx context.Context) *MemDb {
	db := &MemDb{}
	db.BaseDb.defaultLock()
	_ = db.Connect(ctx, "")
	return db
}

// Connect implements Db
func(mdb *MemDb) Connect(ctx context.Context, connStr string) error {
	if mdb.store != nil {
		return nil
	}
	mdb.store = make(map[string][]byte)
	return nil
}

// convert to a supported map key type
func(mdb *MemDb) toHexKey(key []byte) (string, error) {
	k, err := mdb.ToKey(key)
	return hex.EncodeToString(k), err
}

// Get implements Db
func(mdb *MemDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	k, err := mdb.toHexKey(key)
	if err != nil {
		return nil, err
	}
	Logg.TraceCtxf(ctx, "mem get", "k", k)
	v, ok := mdb.store[k]
	if !ok {
		b, _ := hex.DecodeString(k)
		return nil, NewErrNotFound(b)
	}
	return v, nil
}

// Put implements Db
func(mdb *MemDb) Put(ctx context.Context, key []byte, val []byte) error {
	if !mdb.checkPut() {
		return errors.New("unsafe put and safety set")
	}
	k, err := mdb.toHexKey(key)
	if err != nil {
		return err
	}
	mdb.store[k] = val
	Logg.TraceCtxf(ctx, "mem put", "k",  k, "v", val)
	return nil
}

// Close implements Db
func(mdb *MemDb) Close() error {
	return nil
}
