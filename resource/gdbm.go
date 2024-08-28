package resource

import (
	"context"
	"fmt"

	gdbm "github.com/graygnuorg/go-gdbm"
	"git.defalsify.org/vise.git/lang"
)


type gdbmResource struct {
	db *gdbm.Database
	fns map[string]EntryFunc
}

func NewGdbmResource(fp string) *gdbmResource {
	db, err := gdbm.Open(fp, gdbm.ModeReader)
	if err != nil {
		panic(err)
	}
	return &gdbmResource{
		db: db,
	}
}

func ToDbKey(typ uint8, s string, l *lang.Language) []byte {
	k := []byte{typ}
	if l != nil && l.Code != "" {
		s += "_" + l.Code
	}
	return append(k, []byte(s)...)
}


func(dbr *gdbmResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	var ln lang.Language
	v := ctx.Value("Language")
	if v != nil {
		ln = v.(lang.Language)
	}
	k := ToDbKey(FSRESOURCETYPE_TEMPLATE, sym, &ln)
	r, err := dbr.db.Fetch(k)
	if err != nil {
		if err.(*gdbm.GdbmError).Is(gdbm.ErrItemNotFound) {
			k = ToDbKey(FSRESOURCETYPE_TEMPLATE, sym, nil)
			r, err = dbr.db.Fetch(k)
			if err != nil {
				return "", err
			}
		}
	}
	return string(r), nil
}

func(dbr *gdbmResource) GetCode(sym string) ([]byte, error) {
	k := ToDbKey(FSRESOURCETYPE_BIN, sym, nil)
	return dbr.db.Fetch(k)
}

func(dbr *gdbmResource) GetMenu(ctx context.Context, sym string) (string, error) {
	msym := sym + "_menu"
	var ln lang.Language
	v := ctx.Value("Language")
	if v != nil {
		ln = v.(lang.Language)
	}
	k := ToDbKey(FSRESOURCETYPE_TEMPLATE, msym, &ln)
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

func(dbr *gdbmResource) String() string {
	return fmt.Sprintf("gdbm: %v", dbr.db)
}
