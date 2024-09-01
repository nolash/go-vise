package resource

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/db"
)

const (
	resource_max_datatype = db.DATATYPE_STATICLOAD
)

// DbResource is a MenuResource that uses the given db.Db implementation as data retriever.
//
// The DbResource can resolve any db.DATATYPE_* if instructed to do so.
type DbResource struct {
	*MenuResource
	typs uint8
	db db.Db
}

// NewDbResource instantiates a new DbResource
//
// By default it will handle db.DATATYPE_TEPMLATE, db.DATATYPE_MENU and db.DATATYPE_BIN.
func NewDbResource(store db.Db) *DbResource {
	if !store.Safe() {
		logg.Warnf("Db is not safe for use with resource. Make sure it is properly locked before issuing the first retrieval, or it will panic!")
	}
	return &DbResource{
		MenuResource: NewMenuResource(),
		db: store,
		typs: db.DATATYPE_TEMPLATE | db.DATATYPE_MENU | db.DATATYPE_BIN,
	}
}

// Without is a chainable function that disables handling of the given data type.
func(g *DbResource) Without(typ uint8) *DbResource {
	g.typs &= ^typ
	return g
}

// Without is a chainable function that enables handling of the given data type.
func(g *DbResource) With(typ uint8) *DbResource {
	g.typs |= typ
	return g
}

// WithOnly is a chainable convenience function that disables handling of all except the given data type.
func(g *DbResource) WithOnly(typ uint8) *DbResource {
	g.typs = typ
	return g
}

func(g *DbResource) mustSafe() {
	if !g.db.Safe() {
		panic("db unsafe for resource (db.Db.Safe() == false)")
	}
}

// retrieve from underlying db.
func(g *DbResource) fn(ctx context.Context, sym string) ([]byte, error) {
	g.mustSafe()
	return g.db.Get(ctx, []byte(sym))
}

// retrieve from underlying db using a string key.
func(g *DbResource) sfn(ctx context.Context, sym string) (string, error) {
	b, err := g.fn(ctx, sym)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetTemplate implements the Resource interface.
//
// Will fail if support for db.DATATYPE_TEMPLATE has been disabled.
func(g *DbResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	if g.typs & db.DATATYPE_TEMPLATE == 0{
		return "", errors.New("not a template getter")
	}
	g.db.SetPrefix(db.DATATYPE_TEMPLATE)
	return g.sfn(ctx, sym)
}

// GetTemplate implements the Resource interface.
//
// Will fail if support for db.DATATYPE_MENU has been disabled.
func(g *DbResource) GetMenu(ctx context.Context, sym string) (string, error) {
	if g.typs & db.DATATYPE_MENU == 0{
		return "", errors.New("not a menu getter")
	}
	g.db.SetPrefix(db.DATATYPE_MENU)
	qSym := sym + "_menu"
	v, err := g.sfn(ctx, qSym)
	if err != nil {
		if db.IsNotFound(err) {
			logg.TraceCtxf(ctx, "menu unresolved", "sym", sym)
			v = sym
		}
	}
	return v, nil
}

// GetCode implements the Resource interface.
//
// Will fail if support for db.DATATYPE_BIN has been disabled.
func(g *DbResource) GetCode(ctx context.Context, sym string) ([]byte, error) {
	if g.typs & db.DATATYPE_BIN == 0{
		return nil, errors.New("not a code getter")
	}
	g.db.SetPrefix(db.DATATYPE_BIN)
	return g.fn(ctx, sym)
}

// FuncFor implements the Resource interface.
//
// The method will first attempt to resolve using the function registered
// with the MenuResource parent class.
// 
// If no match is found, and if support for db.DATATYPE_STATICLOAD has been enabled,
// an additional lookup will be performed using the underlying db.
func(g *DbResource) FuncFor(ctx context.Context, sym string) (EntryFunc, error) {
	fn, err := g.MenuResource.FuncFor(ctx, sym)
	if err == nil {
		return fn, nil
	}
	if g.typs & db.DATATYPE_STATICLOAD == 0 {
		return nil, errors.New("not a staticload getter")
	}
	g.db.SetPrefix(db.DATATYPE_STATICLOAD)
	b, err := g.fn(ctx, sym)
	if err != nil {
		if !db.IsNotFound(err) {
			return nil, err
		}
		b, err = g.fn(ctx, sym + ".txt")
		if err != nil {
			return nil, err
		}
	}
	return func(ctx context.Context, nodeSym string, input []byte) (Result, error) {
		return Result{
			Content: string(b),
		}, nil
	}, nil
}
