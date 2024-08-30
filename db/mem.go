package db

import (
	"context"
	"fmt"
)

// MemDb is a memory backend implementation of the Db interface.
type MemDb struct {
	BaseDb
	store map[string][]byte
}

// Connect implements Db
func(mdb *MemDb) Connect(ctx context.Context, connStr string) error {
	mdb.store = make(map[string][]byte)
	return nil
}

// convert to a supported map key type
func(mdb *MemDb) toHexKey(key []byte) (string, error) {
	k, err := mdb.ToKey(key)
	return fmt.Sprintf("%x", k), err
}

// Get implements Db
func(mdb *MemDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	k, err := mdb.toHexKey(key)
	if err != nil {
		return nil, err
	}
	v, ok := mdb.store[k]
	if !ok {
		return nil, NewErrNotFound([]byte(k))
	}
	return v, nil
}

// Put implements Db
func(mdb *MemDb) Put(ctx context.Context, key []byte, val []byte) error {
	k, err := mdb.toHexKey(key)
	if err != nil {
		return err
	}
	mdb.store[k] = val
	return nil
}

// Close implements Db
func(mdb *MemDb) Close() error {
	return nil
}
