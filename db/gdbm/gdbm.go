package gdbm

import (
	"context"
	"errors"
	"fmt"
	"os"

	gdbm "github.com/graygnuorg/go-gdbm"

	"git.defalsify.org/vise.git/db"
)

// gdbmDb is a gdbm backend implementation of the Db interface.
type gdbmDb struct {
	*db.DbBase
	conn *gdbm.Database
	readOnly bool
	prefix uint8
	it gdbm.DatabaseIterator
	itBase []byte
}

// Creates a new gdbm backed Db implementation.
func NewGdbmDb() *gdbmDb {
	db := &gdbmDb{
		DbBase: db.NewDbBase(),
	}
	return db
}

// WithReadOnly sets database as read only.
//
// There may exist more than one instance of read-only
// databases to the same file at the same time.
// However, only one single write database.
//
// Readonly cannot be set when creating a new database.
func(gdb *gdbmDb) WithReadOnly() *gdbmDb {
	gdb.readOnly = true
	return gdb
}

// String implements the string interface.
func(gdb *gdbmDb) String() string {
	fn, err := gdb.conn.FileName()
	if err != nil {
		fn = "??"
	}
	return "gdbmdb: " + fn
}

// Connect implements Db
func(gdb *gdbmDb) Connect(ctx context.Context, connStr string) error {
	if gdb.conn != nil {
		logg.WarnCtxf(ctx, "already connected", "conn", gdb.conn)
		return nil
	}
	var db *gdbm.Database
	cfg := gdbm.DatabaseConfig{
		FileName: connStr,
		Flags: gdbm.OF_NOLOCK | gdbm.OF_PREREAD,
		FileMode: 0600,
	}

	_, err := os.Stat(connStr)
	if err != nil {
		if !errors.Is(err.(*os.PathError).Unwrap(), os.ErrNotExist) {
			return fmt.Errorf("db path lookup err: %v", err)
		}
		if gdb.readOnly {
			return fmt.Errorf("cannot open new database readonly")
		}
		cfg.Mode = gdbm.ModeReader | gdbm.ModeWrcreat
	} else {
		cfg.Mode = gdbm.ModeReader
		if !gdb.readOnly {
			cfg.Mode |= gdbm.ModeWriter
		}
	}
	db, err = gdbm.OpenConfig(cfg)
	if err != nil {
		return fmt.Errorf("db open err: %v", err)
	}
	logg.DebugCtxf(ctx, "gdbm connected", "connstr", connStr)
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
	logg.TraceCtxf(ctx, "gdbm get", "key", key, "lk", lk.Default)
	if err != nil {
		if errors.Is(gdbm.ErrItemNotFound, err) {
			return nil, db.NewErrNotFound(key)
		}
		return nil, err
	}
	logg.TraceCtxf(ctx, "gdbm get", "key", key, "lk", lk, "val", v)
	return v, nil
}

// Close implements Db
func(gdb *gdbmDb) Close(ctx context.Context) error {
	logg.TraceCtxf(ctx, "closing gdbm", "path", gdb.conn)
	return gdb.conn.Close()
}
