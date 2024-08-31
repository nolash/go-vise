// Executable gdbm processes a given directory recursively and inserts all template files, menu files and bytecode files into corresponding db.Db entries backed by a gdbm backend.
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
	"strings"

	gdbm "github.com/graygnuorg/go-gdbm"

	"git.defalsify.org/vise.git/db"
)

var (
	binaryPrefix = ".bin"
	menuPrefix = "menu"
	templatePrefix = ""
	scan = make(map[string]string)
	dbg = map[uint8]string{
		db.DATATYPE_BIN: "BIN",
		db.DATATYPE_TEMPLATE: "TEMPLATE",
		db.DATATYPE_MENU: "MENU",
	}
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
	typ = db.DATATYPE_UNKNOWN
	if d.IsDir() {
		return nil
	}
	fx := path.Ext(fp)
	fb := path.Base(fp)
	if (len(fb) == 0) {
		return nil
	}
	if (fb[0] < 0x61 || fb[0] > 0x7A) {
		return nil
	}
	switch fx {
		case binaryPrefix:
			typ = db.DATATYPE_BIN
		case templatePrefix:
			if strings.Contains(fb, "_menu") {
				typ = db.DATATYPE_TEMPLATE
			} else {
				typ = db.DATATYPE_MENU
			}
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
	k := db.ToDbKey(typ, []byte(ft), nil)
	err = sc.db.Store(k, v, true)
	if err != nil {
		return err
	}
	log.Printf("stored key [%s] %x for %s (%s)", dbg[typ], k, fp, ft)
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
