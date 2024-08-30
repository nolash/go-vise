package db

import (
	"context"
	"errors"

	gdbm "github.com/graygnuorg/go-gdbm"
)

type GdbmDb struct {
	conn *gdbm.Database
	prefix uint8
}

func NewGdbmDb() *GdbmDb {
	return &GdbmDb{
		prefix: DATATYPE_USERSTART,
	}
		
}
func(gdb *GdbmDb) Connect(ctx context.Context, connStr string) error {
	db, err := gdbm.Open(connStr, gdbm.ModeWrcreat)
	if err != nil {
		return err
	}
	gdb.conn = db
	return nil
}

// TODO: DRY
func(gdb *GdbmDb) dbKey(sessionId string, key []byte) []byte {
	b := append([]byte(sessionId), 0x2E)
	b = append(b, key...)
	return ToDbKey(gdb.prefix, b, nil)
}

func(gdb *GdbmDb) Put(ctx context.Context, sessionId string, key []byte, val []byte) error {
	k := gdb.dbKey(sessionId, key)
	return gdb.conn.Store(k, val, true)
}

func(gdb *GdbmDb) Get(ctx context.Context, sessionId string, key []byte) ([]byte, error) {
	k := gdb.dbKey(sessionId, key)
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
