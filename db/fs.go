package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type FsDb struct {
	ready bool
	dir string
}

func(fds *FsDb) Connect(connStr string) error {
	fi, err := os.Stat(connStr)
	if err != nil {
		return err
	}
	if !fi.IsDir()  {
		return fmt.Errorf("fs db %s is not a directory", connStr)
	}
	fds.dir = connStr
	return nil
}

func(fsd *FsDb) pathFor(key []byte) string{
	return path.Join(fsd.dir, string(key))
}

func(fsd *FsDb) Get(key []byte) ([]byte, error) {
	fp := fsd.pathFor(key)
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

func(fsd *FsDb) Put(key []byte, val []byte) error {
	fp := fsd.pathFor(key)
	return ioutil.WriteFile(fp, val, 0600)
}
