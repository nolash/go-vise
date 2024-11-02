package mem

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/db"
)

func(mdb *memDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	return nil, errors.New("unimplemented")
}
