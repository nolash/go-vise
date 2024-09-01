package resourcetest

import (
	"context"

	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/db/mem"
	"git.defalsify.org/vise.git/db"
)

type TestResource struct {
	*resource.DbResource
	db db.Db
	ctx context.Context
}

func NewTestResource() *TestResource {
	return newTestResource("")
}

func NewTestFsResource(path string) *TestResource {
	if path == "" {
		panic("empty path")
	}
	return newTestResource(path)
}

func newTestResource(path string) *TestResource {
	var store db.Db
	ctx := context.Background()

	if path == "" {
		mem := mem.NewMemDb()
		mem.SetLock(db.DATATYPE_TEMPLATE, false)
		mem.SetLock(db.DATATYPE_BIN, false)
		mem.SetLock(db.DATATYPE_MENU, false)
		store = mem
	} else {
		fs := mem.NewMemDb()
		fs.SetLock(db.DATATYPE_TEMPLATE, false)
		fs.SetLock(db.DATATYPE_BIN, false)
		fs.SetLock(db.DATATYPE_MENU, false)
		store = fs
	}

	store.Connect(ctx, path)
	rsd := resource.NewDbResource(store)
	rs := &TestResource{
		DbResource: rsd,
		ctx: ctx,
		db: store,
	}
	return rs
}

func(tr *TestResource) AddTemplate(ctx context.Context, key string, val string) error {
	tr.db.SetPrefix(db.DATATYPE_TEMPLATE)
	return tr.db.Put(ctx, []byte(key), []byte(val))
}

func(tr *TestResource) AddBytecode(ctx context.Context, key string, val []byte) error {
	tr.db.SetPrefix(db.DATATYPE_BIN)
	return tr.db.Put(ctx, []byte(key), val)
}

func(tr *TestResource) AddMenu(ctx context.Context, key string, val string) error {
	tr.db.SetPrefix(db.DATATYPE_MENU)
	return tr.db.Put(ctx, []byte(key), []byte(val))
}

func(tr *TestResource) AddFunc(ctx context.Context, key string, fn resource.EntryFunc) {
	tr.AddLocalFunc(key, fn)
}
