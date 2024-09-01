package gdbm

import (
	"context"
	"errors"
	"os"

	gdbm "github.com/graygnuorg/go-gdbm"

	"git.defalsify.org/vise.git/db"
)

// gdbmDb is a gdbm backend implementation of the Db interface.
type gdbmDb struct {
	*db.DbBase
	conn *gdbm.Database
	prefix uint8
}

// Creates a new gdbm backed Db implementation.
func NewGdbmDb() *gdbmDb {
	db := &gdbmDb{
		DbBase: db.NewDbBase(),
	}
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
	if !gdb.CheckPut() {
		return errors.New("unsafe put and safety set")
	}
	lk, err := gdb.ToKey(ctx, key)
	if err != nil {
		return err
	}
	logg.TraceCtxf(ctx, "gdbm put", "key", key, "lk", lk, "val", val)
	if lk.Translation != nil {
		return gdb.conn.Store(lk.Translation, val, true)
	}
	return gdb.conn.Store(lk.Default, val, true)
}

// Get implements Db
func(gdb *gdbmDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	var v []byte
	lk, err := gdb.ToKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if lk.Translation != nil {
		v, err = gdb.conn.Fetch(lk.Translation)
		if err != nil {
			if !errors.Is(gdbm.ErrItemNotFound, err) {
				return nil, err
			}
		}
		return v, nil
	}
	v, err = gdb.conn.Fetch(lk.Default)
	if err != nil {
		if errors.Is(gdbm.ErrItemNotFound, err) {
			return nil, db.NewErrNotFound(key)
		}
		return nil, err
	}
	return v, nil
}

// Close implements Db
func(gdb *gdbmDb) Close() error {
	return gdb.Close()
}
