package postgres

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/db"
)

func(pdb *pgDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	return nil, errors.New("unimplemented")
}
