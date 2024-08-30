package db 

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgDb struct {
	BaseDb
	conn *pgxpool.Pool
	schema string
	prefix uint8
}

func NewPgDb() *PgDb {
	return &PgDb{
		schema: "public",
	}
}

func(pdb *PgDb) WithSchema(schema string) *PgDb {
	pdb.schema = schema
	return pdb
}

func(pdb *PgDb) Connect(ctx context.Context, connStr string) error {
	var err error
	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	pdb.conn = conn
	return pdb.prepare(ctx)
}

func(pdb *PgDb) prepare(ctx context.Context) error {
	tx, err := pdb.conn.Begin(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
//	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kv_vise_domain (
//		id SERIAL PRIMARY KEY,
//		name VARCHAR(256) NOT NULL
//	);
//`, pdb.schema)
//	_, err = tx.Exec(ctx, query)
//	if err != nil {
//		tx.Rollback(ctx)
//		return err
//	}
//
//	query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kv_vise (
//		id SERIAL NOT NULL,
//		domain_id INT NOT NULL,
//		key VARCHAR(256) NOT NULL,
//		value BYTEA NOT NULL,
//		constraint fk_domain
//			FOREIGN KEY (domain_id)
//			REFERENCES %s.kv_vise_domain(id)
//	);
//`, pdb.schema, pdb.schema)
//	_, err = tx.Exec(ctx, query)
//	if err != nil {
//		tx.Rollback(ctx)
//		return err
//	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kv_vise (
		id SERIAL NOT NULL,
		key BYTEA NOT NULL UNIQUE,
		value BYTEA NOT NULL
	);
`, pdb.schema)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		//if !errors.Is(pgx.ErrTxCommitRollback) {
			tx.Rollback(ctx)
			return err
		//}
	}
	return nil
}

func(pdb *PgDb) Put(ctx context.Context, sessionId string, key []byte, val []byte) error {
	k, err := pdb.ToKey(sessionId, key)
	if err != nil {
		return err
	}
	tx, err := pdb.conn.Begin(ctx)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("INSERT INTO %s.kv_vise (key, value) VALUES ($1, $2) ON CONFLICT(key) DO UPDATE SET value = $2;", pdb.schema)
	_, err = tx.Exec(ctx, query, k, val)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	tx.Commit(ctx)
	return nil
}

func(pdb *PgDb) Get(ctx context.Context, sessionId string, key []byte) ([]byte, error) {
	k, err := pdb.ToKey(sessionId, key)
	if err != nil {
		return nil, err
	}
	tx, err := pdb.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT value FROM %s.kv_vise WHERE key = $1", pdb.schema)
	rs, err := tx.Query(ctx, query, k)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	if !rs.Next() {
		return nil, NewErrNotFound(k)

	}
	r := rs.RawValues()
	b := r[0]
	return b, nil
}

func(pdb *PgDb) Close() error {
	pdb.Close()
	return nil
}
