package db

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type FsDb struct {
	ready bool
	dir string
}

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

func(fdb *FsDb) Get(ctx context.Context, sessionId string, key []byte) ([]byte, error) {
	fp := fdb.pathFor(sessionId, key)
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func(fdb *FsDb) Put(ctx context.Context, sessionId string, key []byte, val []byte) error {
	fp := fdb.pathFor(sessionId, key)
	return ioutil.WriteFile(fp, val, 0600)
}

func(fdb *FsDb) pathFor(sessionId string, key []byte) string{
	k := sessionId + "." + string(key)
	kb := ToDbKey(DATATYPE_USERSTART, k, nil)
	kb[0] += 30
	return path.Join(fdb.dir, string(kb))
}
