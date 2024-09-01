package resource

import (
	"context"
	"errors"
	"fmt"

	"git.defalsify.org/vise.git/db"
)

const (
	resource_max_datatype = db.DATATYPE_STATICLOAD
)

// DbResource is a MenuResource that uses the given db.Db implementation as data retriever.
type DbResource struct {
	*MenuResource
	typs uint8
	db db.Db
}

// NewDbFuncGetter instantiates a new DbResource
func NewDbResource(store db.Db, typs... uint8) (*DbResource, error) {
	var v uint8
	g := &DbResource{
		MenuResource: NewMenuResource(),
		db: store,
	}
	for _, v = range(typs) {
		if v > resource_max_datatype {
			return nil, fmt.Errorf("datatype %d is not a resource", v)	
		}
		g.typs |= v
	}
	return g, nil
}

func(g *DbResource) fn(ctx context.Context, sym string) ([]byte, error) {
	return g.db.Get(ctx, []byte(sym))
}

func(g *DbResource) sfn(ctx context.Context, sym string) (string, error) {
	b, err := g.fn(ctx, sym)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetTemplate implements the Resource interface.
func(g *DbResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	if g.typs & db.DATATYPE_TEMPLATE == 0{
		return "", errors.New("not a template getter")
	}
	g.db.SetPrefix(db.DATATYPE_TEMPLATE)
	return g.sfn(ctx, sym)
}

// GetTemplate implements the Resource interface.
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

// GetTemplate implements the Resource interface.
func(g *DbResource) GetCode(ctx context.Context, sym string) ([]byte, error) {
	if g.typs & db.DATATYPE_BIN == 0{
		return nil, errors.New("not a code getter")
	}
	g.db.SetPrefix(db.DATATYPE_BIN)
	return g.fn(ctx, sym)
}

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
		return nil, err
	}
	return func(ctx context.Context, nodeSym string, input []byte) (Result, error) {
		return Result{
			Content: string(b),
		}, nil
	}, nil
}
