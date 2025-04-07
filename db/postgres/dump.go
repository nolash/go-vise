package postgres

import (
	"context"
	"fmt"

	"git.defalsify.org/vise.git/db"
)

// Dump implements Db.
func (pdb *pgDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	tx, err := pdb.conn.BeginTx(ctx, defaultTxOptions)
	if err != nil {
		return nil, err
	}

	pdb.SetLanguage(nil)
	lk, err := pdb.ToKey(ctx, key)
	if err != nil {
		return nil, err
	}
	k := lk.Default

	query := fmt.Sprintf("SELECT key, value FROM %s.kv_vise WHERE key >= $1", pdb.schema)
	rs, err := tx.Query(ctx, query, k)
	if err != nil {
		logg.Debugf("query fail", "err", err)
		tx.Rollback(ctx)
		return nil, err
	}
	defer tx.Commit(ctx)

	if rs.Next() {
		var kk []byte
		var vv []byte
		err = rs.Scan(&kk, &vv)
		if err != nil {
			return nil, err
		}
		pdb.it = rs
		pdb.itBase = k
		kk, err = pdb.DecodeKey(ctx, kk)
		if err != nil {
			return nil, err
		}
		return db.NewDumper(pdb.dumpFunc).WithClose(pdb.closeFunc).WithFirst(kk, vv), nil
	}

	return nil, db.NewErrNotFound(k)
}

func (pdb *pgDb) dumpFunc(ctx context.Context) ([]byte, []byte) {
	var kk []byte
	var vv []byte
	if !pdb.it.Next() {
		logg.DebugCtxf(ctx, "no more data in pg iterator")
		pdb.it = nil
		pdb.itBase = nil
		return nil, nil
	}
	err := pdb.it.Scan(&kk, &vv)
	if err != nil {
		return nil, nil
	}
	k, err := pdb.DecodeKey(ctx, kk)
	if err != nil {
		return nil, nil
	}
	return k, vv
}

func (pdb *pgDb) closeFunc() error {
	if pdb.it != nil {
		pdb.it.Close()
		pdb.it = nil
	}
	return nil
}
