package testdata

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/logging"
)

var (
	ctx = context.Background()
	store = db.NewFsDb()
	out = outNew
	logg = logging.NewVanilla().WithDomain("testdata")
)

type echoFunc struct {
	v string
}

func(e *echoFunc) get(ctx context.Context, nodeSym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: e.v,
	}, nil
}

func outNew(sym string, b []byte, tpl string, data map[string]string) error {
	logg.Debugf("testdata out", "sym", sym)
	store.SetPrefix(db.DATATYPE_TEMPLATE)
	err := store.Put(ctx, []byte(sym), []byte(tpl))
	if err != nil {
		return err
	}
	store.SetPrefix(db.DATATYPE_BIN)
	err = store.Put(ctx, []byte(sym), b)
	if err != nil {
		return err
	}
	store.SetPrefix(db.DATATYPE_STATICLOAD)
	for k, v := range data {
		logg.Debugf("testdata out staticload", "sym", sym, "k", k, "v", v)
		err = store.Put(ctx, []byte(k), []byte(v))
		if err != nil {
			return err
		}
	}
	return nil
}

func generate() error {
	err := os.MkdirAll(DataDir, 0755)
	if err != nil {
		return err
	}
	store = db.NewFsDb()
	store.Connect(ctx, DataDir)
	store.SetLock(db.DATATYPE_TEMPLATE, false)
	store.SetLock(db.DATATYPE_BIN, false)
	store.SetLock(db.DATATYPE_MENU, false)
	store.SetLock(db.DATATYPE_STATICLOAD, false)

	fns := []genFunc{root, foo, bar, baz, long, lang, defaultCatch}
	for _, fn := range fns {
		err = fn()
		if err != nil {
			return err
		}
	}
	return nil
}

// Generate outputs bytecode, templates and content symbols to a temporary directory.
//
// This directory can in turn be used as data source for the the resource.FsResource object.
func Generate() (string, error) {
	dir, err := ioutil.TempDir("", "vise_testdata_")
	if err != nil {
		return "", err
	}
	DataDir = dir
	dirLock = true
	err = generate()
	return dir, err
}


// Generate outputs bytecode, templates and content symbols to a specified directory.
//
// The directory must exist, and must not have been used already in the same code execution.
//
// This directory can in turn be used as data source for the the resource.FsResource object.
func GenerateTo(dir string) error {
	if dirLock {
		return fmt.Errorf("directory already overridden")
	}
	DataDir = dir
	dirLock = true
	return generate()
}
