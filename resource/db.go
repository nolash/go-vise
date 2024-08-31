package resource

import (
	"context"
	"errors"
	"fmt"

	"git.defalsify.org/vise.git/db"
)

const (
	resource_max_datatype = db.DATATYPE_TEMPLATE
)

type dbResource struct {
	MenuResource
	typs uint8
	db db.Db
}

// NewDbFuncGetter returns a MenuResource that uses the given db.Db implementation as data retriever.
func NewDbResource(store db.Db, typs... uint8) (*dbResource, error) {
	var v uint8
	g := &dbResource{
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

func(g *dbResource) fn(ctx context.Context, sym string) ([]byte, error) {
	return g.db.Get(ctx, []byte(sym))
}

func(g *dbResource) sfn(ctx context.Context, sym string) (string, error) {
	b, err := g.fn(ctx, sym)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetTemplate implements the Resource interface.
func(g *dbResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	if g.typs & db.DATATYPE_TEMPLATE == 0{
		return "", errors.New("not a template getter")
	}
	g.db.SetPrefix(db.DATATYPE_TEMPLATE)
	return g.sfn(ctx, sym)
}

// GetTemplate implements the Resource interface.
func(g *dbResource) GetMenu(ctx context.Context, sym string) (string, error) {
	if g.typs & db.DATATYPE_MENU == 0{
		return "", errors.New("not a menu getter")
	}
	g.db.SetPrefix(db.DATATYPE_MENU)
	return g.sfn(ctx, sym)
}

// GetTemplate implements the Resource interface.
func(g *dbResource) GetCode(ctx context.Context, sym string) ([]byte, error) {
	if g.typs & db.DATATYPE_BIN == 0{
		return nil, errors.New("not a code getter")
	}
	g.db.SetPrefix(db.DATATYPE_BIN)
	return g.fn(ctx, sym)
}
