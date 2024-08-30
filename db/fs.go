package db

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// FsDb is a pure filesystem backend implementation if the Db interface.
type FsDb struct {
	BaseDb
	dir string
}

// Connect implements Db
func(fdb *FsDb) Connect(ctx context.Context, connStr string) error {
	fi, err := os.Stat(connStr)
	if err != nil {
		return err
	}
	if !fi.IsDir()  {
		return fmt.Errorf("fs db %s is not a directory", connStr)
	}
	fdb.dir = connStr
	return nil
}

// Get implements Db
func(fdb *FsDb) Get(ctx context.Context, key []byte) ([]byte, error) {
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
func(fdb *FsDb) Put(ctx context.Context, key []byte, val []byte) error {
	fp, err := fdb.pathFor(key)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fp, val, 0600)
}

// Close implements Db
func(fdb *FsDb) Close() error {
	return nil
}	

// create a key safe for the filesystem
func(fdb *FsDb) pathFor(key []byte) (string, error) {
	kb, err := fdb.ToKey(key)
	if err != nil {
		return "", err
	}
	kb[0] += 30
	return path.Join(fdb.dir, string(kb)), nil
}
