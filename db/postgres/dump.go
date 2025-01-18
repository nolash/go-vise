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

	pdb.SetPrefix(db.DATATYPE_USERDATA)
	pdb.SetLanguage(nil)
	lk, err := pdb.ToKey(ctx, key)
	if err != nil {
		return nil, err
	}
	k := lk.Default

	query := fmt.Sprintf("SELECT key, value FROM %s.kv_vise WHERE key >= $1", pdb.schema)
	logg.TraceCtxf(ctx, "getkey", "q", query, "key", k)
	rs, err := tx.Query(ctx, query, k)
	if err != nil {
		logg.Debugf("query fail", "err", err)
		tx.Rollback(ctx)
		return nil, err
	}
	defer tx.Commit(ctx)
	//defer rs.Close()

	if rs.Next() {
		var kk []byte
		var vv []byte
//		r, err := rs.Values()
//		if err != nil {
//			return nil, err
//		}
		err = rs.Scan(&kk, &vv)
		if err != nil {
			return nil, err
		}
		//tx.Rollback(ctx)
		//tx.Commit(ctx)
		pdb.it = rs
		pdb.itBase = k
		kk, err = pdb.DecodeKey(ctx, kk)
		logg.Debugf("pg decode", "k", kk, "o", k, "err", err, "vv", vv)
		if err != nil {
			return nil, err
		}
		return db.NewDumper(pdb.dumpFunc).WithClose(pdb.closeFunc).WithFirst(kk, vv), nil
	}

	return nil, db.NewErrNotFound(k)
}

func(pdb *pgDb) dumpFunc(ctx context.Context) ([]byte, []byte) {
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
	//r := pdb.it.RawValues()
	//k, err := pdb.DecodeKey(ctx, r[0])
	k, err := pdb.DecodeKey(ctx, kk)
	if err != nil {
		return nil, nil
	}
	logg.Debugf("pg decode dump", "k", kk, "o", k, "err", err, "vv", vv)
	return k, vv
}

func(pdb *pgDb) closeFunc() error {
	if pdb.it != nil {
		pdb.it.Close()
		pdb.it = nil
	}
	return nil
}
