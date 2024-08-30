package db

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
)

// fsDb is a pure filesystem backend implementation if the Db interface.
type fsDb struct {
	baseDb
	dir string
}

func NewFsDb() *fsDb {
	db := &fsDb{}
	db.baseDb.defaultLock()
	return db
}

// Connect implements Db
func(fdb *fsDb) Connect(ctx context.Context, connStr string) error {
	if fdb.dir != "" {
		return nil
	}
	err := os.MkdirAll(connStr, 0700)
	if err != nil {
		return err
	}
	fdb.dir = connStr
	return nil
}

// Get implements Db
func(fdb *fsDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	fp, err := fdb.pathFor(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fp)
	if err != nil {
		return nil, NewErrNotFound([]byte(fp))
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Put implements Db
func(fdb *fsDb) Put(ctx context.Context, key []byte, val []byte) error {
	if !fdb.checkPut() {
		return errors.New("unsafe put and safety set")
	}
	fp, err := fdb.pathFor(key)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fp, val, 0600)
}

// Close implements Db
func(fdb *fsDb) Close() error {
	return nil
}	

// create a key safe for the filesystem
func(fdb *fsDb) pathFor(key []byte) (string, error) {
	kb, err := fdb.ToKey(key)
	if err != nil {
		return "", err
	}
	kb[0] += 0x30
	return path.Join(fdb.dir, string(kb)), nil
}
