package db

import (
	"context"
	"errors"

	gdbm "github.com/graygnuorg/go-gdbm"
)

type GdbmDb struct {
	BaseDb
	conn *gdbm.Database
	prefix uint8
}

func(gdb *GdbmDb) Connect(ctx context.Context, connStr string) error {
	db, err := gdbm.Open(connStr, gdbm.ModeWrcreat)
	if err != nil {
		return err
	}
	gdb.conn = db
	return nil
}

func(gdb *GdbmDb) Put(ctx context.Context, sessionId string, key []byte, val []byte) error {
	k, err := gdb.ToKey(sessionId, key)
	if err != nil {
		return err
	}
	return gdb.conn.Store(k, val, true)
}

func(gdb *GdbmDb) Get(ctx context.Context, sessionId string, key []byte) ([]byte, error) {
	k, err := gdb.ToKey(sessionId, key)
	if err != nil {
		return nil, err
	}
	v, err := gdb.conn.Fetch(k)
	if err != nil {
		if errors.Is(gdbm.ErrItemNotFound, err) {
			return nil, NewErrNotFound(k)
		}
		return nil, err
	}
	return v, nil
}

func(gdb *GdbmDb) Close() error {
	return gdb.Close()
}
