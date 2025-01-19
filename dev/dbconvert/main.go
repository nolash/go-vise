// Executable dbconvert processes a given directory recursively and inserts all legacy template files, menu files and bytecode files into corresponding db.Db entries of the chosen backend.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"path"
	"strings"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"
	"git.defalsify.org/vise.git/logging"
)

var (
	binaryPrefix = ".bin"
	menuPrefix = "menu"
	staticloadPrefix = ".txt"
	templatePrefix = ""
	scan = make(map[string]string)
	logg = logging.NewVanilla()
	dbg = map[uint8]string{
		db.DATATYPE_BIN: "BIN",
		db.DATATYPE_TEMPLATE: "TEMPLATE",
		db.DATATYPE_MENU: "MENU",
		db.DATATYPE_STATICLOAD: "STATICLOAD",
	}
)

type scanner struct {
	ctx context.Context
	db db.Db
}

func newScanner(ctx context.Context, db db.Db) (*scanner, error) {
	return &scanner{
		ctx: ctx,
		db: db,
	}, nil
}

func(sc *scanner) Close() error {
	return sc.db.Close(sc.ctx)
}

func(sc *scanner) Scan(fp string, d fs.DirEntry, err error) error { 
	if err != nil {
		return err
	}
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
	sc.db.SetPrefix(db.DATATYPE_UNKNOWN)
	switch fx {
		case binaryPrefix:
			sc.db.SetPrefix(db.DATATYPE_BIN)
			//typ = db.DATATYPE_BIN
		case templatePrefix:
			if strings.Contains(fb, "_menu") {
				sc.db.SetPrefix(db.DATATYPE_TEMPLATE)
				//typ = db.DATATYPE_TEMPLATE
			} else {
				sc.db.SetPrefix(db.DATATYPE_MENU)
				//typ = db.DATATYPE_MENU
			}
		case staticloadPrefix:
			sc.db.SetPrefix(db.DATATYPE_STATICLOAD)
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

	logg.TraceCtxf(sc.ctx, "put record", "fx", fx, "fb", fb)
	ft := fb[:len(fb)-len(fx)]
	err = sc.db.Put(sc.ctx, []byte(ft), v)
	if err != nil {
		return err
	}
	//k := db.ToDbKey(typ, []byte(ft), nil)
	//err = sc.db.Store(k, v, true)
	//if err != nil {
	//	return err
	//}
	return nil
}

func main() {
	var store db.Db
	var err error
	var dir string
	var dbPath string
	var dbFile string
	var dbBackend string
	flag.StringVar(&dbPath, "d", "", "output directory")
	flag.StringVar(&dbBackend, "backend", "gdbm", "db backend. valid choices are: gdbm (default), fs")
	flag.Parse()

	ctx := context.Background()
	switch dbBackend {
	case "gdbm":
		store = gdbmdb.NewGdbmDb()
		dbFile = "vise_resources.gdbm"
	case "fs":
		store = fsdb.NewFsDb()
	}

	dir = flag.Arg(0)

	if dbPath == "" {
		dbPath, err = os.MkdirTemp(dir, "vise-dbconvert-*")
	} else {
		err = os.MkdirAll(dir, 0700)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output dir")
		os.Exit(1)
	}
	if dbFile != "" {
		dbPath = path.Join(dbPath, dbFile)
	}
	if dir == dbPath {
		fmt.Fprintf(os.Stderr, "input and output dir cannot be the same")
	}

	err = store.Connect(ctx, dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to output db: %s", err)
		os.Exit(1)
	}
	
	store.SetLock(db.DATATYPE_BIN, false)
	store.SetLock(db.DATATYPE_TEMPLATE, false)
	store.SetLock(db.DATATYPE_MENU, false)
	store.SetLock(db.DATATYPE_STATICLOAD, false)

	o, err := newScanner(ctx, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open scanner")
		os.Exit(1)
	}
	err = filepath.WalkDir(dir, o.Scan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to process input: %s", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, dbPath)
}
