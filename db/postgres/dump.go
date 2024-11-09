package postgres

import (
	"fmt"
	"context"

	"git.defalsify.org/vise.git/db"
)

func(pdb *pgDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	tx, err := pdb.conn.BeginTx(ctx, defaultTxOptions)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT key, value FROM %s.kv_vise WHERE key >= $1 AND key < $2", pdb.schema)
	rs, err := tx.Query(ctx, query, key, key[0])
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	defer rs.Close()

	if rs.Next() {
		r := rs.RawValues()
		tx.Commit(ctx)
		//tx.Rollback(ctx)
		pdb.it = rs
		pdb.itBase = key
		return db.NewDumper(pdb.dumpFunc).WithFirst(r[0], r[1]), nil
	}

	return nil, db.NewErrNotFound(key)
}

func(pdb *pgDb) dumpFunc(ctx context.Context) ([]byte, []byte) {
	if !pdb.it.Next() {
		pdb.it = nil
		pdb.itBase = nil
		return nil, nil
	}
	r := pdb.it.RawValues()
	return r[0], r[1]
}
