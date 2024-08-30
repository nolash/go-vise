package db 

import (
	"context"
	"fmt"
//	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
//	pgx "github.com/jackc/pgx/v5"
)

type PgDb struct {
	conn *pgxpool.Pool
	schema string
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
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kv_vise_domain (
		id SERIAL PRIMARY KEY,
		name VARCHAR(256) NOT NULL
	);
`, pdb.schema)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kv_vise (
		id SERIAL NOT NULL,
		domain_id INT NOT NULL,
		key VARCHAR(256) NOT NULL,
		value BYTEA NOT NULL,
		constraint fk_domain
			FOREIGN KEY (domain_id)
			REFERENCES %s.kv_vise_domain(id)
	);
`, pdb.schema, pdb.schema)
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

func(pdb *PgDb) Put(ctx context.Context, key []byte, val []byte) error {
	return nil
}

func(pdb *PgDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	return nil, nil
}
