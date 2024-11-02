package fs

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/db"
)

func(fdb *fsDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	return nil, errors.New("unimplemented")
}
