package db

import (
	"context"
	"errors"
	"os"

	gdbm "github.com/graygnuorg/go-gdbm"
)

// gdbmDb is a gdbm backend implementation of the Db interface.
type gdbmDb struct {
	baseDb
	conn *gdbm.Database
	prefix uint8
}

// Creates a new gdbm backed Db implementation.
func NewGdbmDb() *gdbmDb {
	db := &gdbmDb{}
	db.baseDb.defaultLock()
	return db
}

// Connect implements Db
func(gdb *gdbmDb) Connect(ctx context.Context, connStr string) error {
	if gdb.conn != nil {
		panic("already connected")
	}
	var db *gdbm.Database
	_, err := os.Stat(connStr)
	if err != nil {
		if !errors.Is(os.ErrNotExist, err) {
			return err
		}
		db, err = gdbm.Open(connStr, gdbm.ModeWrcreat)
	} else {
		db, err = gdbm.Open(connStr, gdbm.ModeWriter | gdbm.ModeReader)
	}

	if err != nil {
		return err
	}
	gdb.conn = db
	return nil
}

// Put implements Db
func(gdb *gdbmDb) Put(ctx context.Context, key []byte, val []byte) error {
	if !gdb.checkPut() {
		return errors.New("unsafe put and safety set")
	}
	k, err := gdb.ToKey(key)
	if err != nil {
		return err
	}
	return gdb.conn.Store(k, val, true)
}

// Get implements Db
func(gdb *gdbmDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	k, err := gdb.ToKey(key)
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

// Close implements Db
func(gdb *gdbmDb) Close() error {
	return gdb.Close()
}
