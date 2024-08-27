package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"path"

	gdbm "github.com/graygnuorg/go-gdbm"
)

var (
	binaryPrefix = ".bin"
	templatePrefix = ""
	scan = make(map[string]string)
)

const (
	RESOURCETYPE_UNKNOWN = iota
	RESOURCETYPE_BIN
	RESOURCETYPE_TEMPLATE
)

type scanner struct {
	db *gdbm.Database
}

func newScanner(fp string) (*scanner, error) {
	db, err := gdbm.Open(fp, gdbm.ModeNewdb)
	if err != nil {
		return nil, err
	}
	return &scanner{
		db: db,
	}, nil
}

func(sc *scanner) Close() error {
	return sc.db.Close()
}

func(sc *scanner) Scan(fp string, d fs.DirEntry, err error) error { 
	if err != nil {
		return err
	}
	typ := RESOURCETYPE_UNKNOWN
	if d.IsDir() {
		return nil
	}
	fx := path.Ext(fp)
	fb := path.Base(fp)
	switch fx {
		case binaryPrefix:
			typ = RESOURCETYPE_BIN
		case templatePrefix:
			typ = RESOURCETYPE_TEMPLATE
		default:
			log.Printf("skip foreign file: %s", fp)
			return nil
	}
	f, err := os.Open(fp)
	defer f.Close()
	if err != nil{
		return err
	}
	v, err := io.ReadAll(f)
	if err != nil{
		return err
	}

	ft := path.Base(fb)
	k := []byte{uint8(typ)}
	k = append(k, []byte(ft)...)
	return sc.db.Store(k, v, true)
}

func main() {
	var dir string
	var dbPath string
	flag.StringVar(&dbPath, "d", "vise.gdbm", "database file path")
	flag.Parse()

	dir = flag.Arg(0)

	o, err := newScanner(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open scanner")
		os.Exit(1)
	}
	err = filepath.WalkDir(dir, o.Scan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open scanner")
		os.Exit(1)
	}
}
