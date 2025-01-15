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

	k := append([]byte{db.DATATYPE_USERDATA}, key...)

	query := fmt.Sprintf("SELECT key, value FROM %s.kv_vise WHERE key >= $1", pdb.schema)
	logg.TraceCtxf(ctx, "getkey", "q", query, "key", k)
	rs, err := tx.Query(ctx, query, k)
	if err != nil {
		logg.Debugf("query fail", "err", err)
		tx.Rollback(ctx)
		return nil, err
	}
	//defer rs.Close()

	if rs.Next() {
		r := rs.RawValues()
		//tx.Rollback(ctx)
		tx.Commit(ctx)
		pdb.it = rs
		pdb.itBase = k
		return db.NewDumper(pdb.dumpFunc).WithClose(pdb.closeFunc).WithFirst(r[0][1:], r[1]), nil
	}

	return nil, db.NewErrNotFound(k)
}

func(pdb *pgDb) dumpFunc(ctx context.Context) ([]byte, []byte) {
	if !pdb.it.Next() {
		logg.DebugCtxf(ctx, "no more data in pg iterator")
		pdb.it = nil
		pdb.itBase = nil
		return nil, nil
	}
	r := pdb.it.RawValues()
	return r[0][1:], r[1]
}

func(pdb *pgDb) closeFunc() error {
	if pdb.it != nil {
		pdb.it.Close()
		pdb.it = nil
	}
	return nil
}
