package db

import (
	"context"
	"fmt"
)

type MemDb struct {
	BaseDb
	store map[string][]byte
}

func(mdb *MemDb) Connect(ctx context.Context, connStr string) error {
	mdb.store = make(map[string][]byte)
	return nil
}

func(mdb *MemDb) toHexKey(key []byte) (string, error) {
	k, err := mdb.ToKey(key)
	return fmt.Sprintf("%x", k), err
}

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

func(mdb *MemDb) Put(ctx context.Context, key []byte, val []byte) error {
	k, err := mdb.toHexKey(key)
	if err != nil {
		return err
	}
	mdb.store[k] = val
	return nil
}

func(mdb *MemDb) Close() error {
	return nil
}
