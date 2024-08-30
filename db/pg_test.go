package db

import (
	"context"
	"testing"
)

func TestCreate(t *testing.T) {
	db := NewPgDb().WithSchema("govise")
	ctx := context.Background()
	err := db.Connect(ctx, "postgres://vise:esiv@localhost:5432/visedb")
	if err != nil {
		t.Fatal(err)
	}
}
