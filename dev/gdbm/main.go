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

	"git.defalsify.org/vise.git/resource"
)

var (
	binaryPrefix = ".bin"
	templatePrefix = ""
	scan = make(map[string]string)
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
	var typ uint8
	if err != nil {
		return err
	}
	typ = resource.FSRESOURCETYPE_UNKNOWN
	if d.IsDir() {
		return nil
	}
	fx := path.Ext(fp)
	fb := path.Base(fp)
	switch fx {
		case binaryPrefix:
			typ = resource.FSRESOURCETYPE_BIN
		case templatePrefix:
			typ = resource.FSRESOURCETYPE_TEMPLATE
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

	log.Printf("fx fb %s %s", fx, fb)
	ft := fb[:len(fb)-len(fx)]
	k := resource.ToDbKey(typ, ft, nil)
	err = sc.db.Store(k, v, true)
	if err != nil {
		return err
	}
	log.Printf("stored key %x for %s (%s)", k, fp, ft)
	return nil
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
