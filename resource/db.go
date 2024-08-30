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

type dbGetter struct {
	typs uint8
	db db.Db
}

func NewDbFuncGetter(store db.Db, typs... uint8) (*dbGetter, error) {
	var v uint8
	g := &dbGetter{
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

func(g *dbGetter) fn(ctx context.Context, sym string) ([]byte, error) {
	return g.db.Get(ctx, []byte(sym))
}

func(g *dbGetter) sfn(ctx context.Context, sym string) (string, error) {
	b, err := g.fn(ctx, sym)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func(g *dbGetter) GetTemplate(ctx context.Context, sym string) (string, error) {
	if g.typs & db.DATATYPE_TEMPLATE == 0{
		return "", errors.New("not a template getter")
	}
	g.db.SetPrefix(db.DATATYPE_TEMPLATE)
	return g.sfn(ctx, sym)
}

func(g *dbGetter) GetMenu(ctx context.Context, sym string) (string, error) {
	if g.typs & db.DATATYPE_MENU == 0{
		return "", errors.New("not a menu getter")
	}
	g.db.SetPrefix(db.DATATYPE_MENU)
	return g.sfn(ctx, sym)
}

func(g *dbGetter) GetCode(ctx context.Context, sym string) ([]byte, error) {
	if g.typs & db.DATATYPE_BIN == 0{
		return nil, errors.New("not a code getter")
	}
	g.db.SetPrefix(db.DATATYPE_BIN)
	return g.fn(ctx, sym)
}
