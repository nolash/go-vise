package resource

import (
	"context"
	"fmt"

	gdbm "github.com/graygnuorg/go-gdbm"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/db"
)

type gdbmResource struct {
	db *gdbm.Database
	fns map[string]EntryFunc
}

func NewGdbmResource(fp string) *gdbmResource {
	gdb, err := gdbm.Open(fp, gdbm.ModeReader)
	if err != nil {
		panic(err)
	}
	return NewGdbmResourceFromDatabase(gdb)
}

func NewGdbmResourceFromDatabase(gdb *gdbm.Database) *gdbmResource {
	return &gdbmResource{
		db: gdb,
	}
}

func(dbr *gdbmResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	var ln lang.Language
	v := ctx.Value("Language")
	if v != nil {
		ln = v.(lang.Language)
	}
	k := db.ToDbKey(db.DATATYPE_TEMPLATE, []byte(sym), &ln)
	r, err := dbr.db.Fetch(k)
	if err != nil {
		if err.(*gdbm.GdbmError).Is(gdbm.ErrItemNotFound) {
			k = db.ToDbKey(db.DATATYPE_TEMPLATE, []byte(sym), nil)
			r, err = dbr.db.Fetch(k)
			if err != nil {
				return "", err
			}
		}
	}
	return string(r), nil
}

func(dbr *gdbmResource) GetCode(sym string) ([]byte, error) {
	k := db.ToDbKey(db.DATATYPE_BIN, []byte(sym), nil)
	return dbr.db.Fetch(k)
}

func(dbr *gdbmResource) GetMenu(ctx context.Context, sym string) (string, error) {
	msym := sym + "_menu"
	var ln lang.Language
	v := ctx.Value("Language")
	if v != nil {
		ln = v.(lang.Language)
	}
	k := db.ToDbKey(db.DATATYPE_TEMPLATE, []byte(msym), &ln)
	r, err := dbr.db.Fetch(k)
	if err != nil {
		if err.(*gdbm.GdbmError).Is(gdbm.ErrItemNotFound) {
			return sym, nil
		}
		return "", err
	}
	return string(r), nil

}

func(dbr gdbmResource) FuncFor(sym string) (EntryFunc, error) {
	fn, ok := dbr.fns[sym]
	if !ok {
		return nil, fmt.Errorf("function %s not found", sym)
	}
	return fn, nil
}

func(dbr *gdbmResource) AddLocalFunc(sym string, fn EntryFunc) {
	if dbr.fns == nil {
		dbr.fns = make(map[string]EntryFunc)
	}
	dbr.fns[sym] = fn
}

// String implements the String interface.
func(dbr *gdbmResource) String() string {
	return fmt.Sprintf("gdbm: %v", dbr.db)
}
