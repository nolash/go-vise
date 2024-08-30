package db 

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PgDb is a Postgresql backend implementation of the Db interface.
type PgDb struct {
	BaseDb
	conn *pgxpool.Pool
	schema string
	prefix uint8
}

// NewPgDb creates a new PgDb reference.
func NewPgDb() *PgDb {
	return &PgDb{
		schema: "public",
	}
}

// WithSchema sets the Postgres schema to use for the storage table.
func(pdb *PgDb) WithSchema(schema string) *PgDb {
	pdb.schema = schema
	return pdb
}

// Connect implements Db.
func(pdb *PgDb) Connect(ctx context.Context, connStr string) error {
	var err error
	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	pdb.conn = conn
	return pdb.prepare(ctx)
}

// Put implements Db.
func(pdb *PgDb) Put(ctx context.Context, key []byte, val []byte) error {
	if !pdb.checkPut() {
		return errors.New("unsafe put and safety set")
	}
	k, err := pdb.ToKey(key)
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

// Get implements Db.
func(pdb *PgDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	k, err := pdb.ToKey(key)
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

// Close implements Db.
func(pdb *PgDb) Close() error {
	pdb.Close()
	return nil
}

// set up table
func(pdb *PgDb) prepare(ctx context.Context) error {
	tx, err := pdb.conn.Begin(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
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
		tx.Rollback(ctx)
		return err
	}
	return nil
}
