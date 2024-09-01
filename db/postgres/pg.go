package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"git.defalsify.org/vise.git/db"
)

// pgDb is a Postgres backend implementation of the Db interface.
type pgDb struct {
	*db.DbBase
	conn *pgxpool.Pool
	schema string
	prefix uint8
}

// NewpgDb creates a new Postgres backed Db implementation.
func NewPgDb() *pgDb {
	db := &pgDb{
		DbBase: db.NewDbBase(),
		schema: "public",
	}
	return db
}

// WithSchema sets the Postgres schema to use for the storage table.
func(pdb *pgDb) WithSchema(schema string) *pgDb {
	pdb.schema = schema
	return pdb
}

// Connect implements Db.
func(pdb *pgDb) Connect(ctx context.Context, connStr string) error {
	if pdb.conn != nil {
		logg.WarnCtxf(ctx, "already connected", "conn", pdb.conn)
		panic("already connected")
	}
	var err error
	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	pdb.conn = conn
	return pdb.prepare(ctx)
}

// Put implements Db.
func(pdb *pgDb) Put(ctx context.Context, key []byte, val []byte) error {
	if !pdb.CheckPut() {
		return errors.New("unsafe put and safety set")
	}
	k, err := pdb.ToKey(ctx, key)
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
func(pdb *pgDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	lk, err := pdb.ToKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if lk.Translation != nil {
		tx, err := pdb.conn.Begin(ctx)
		if err != nil {
			return nil, err
		}
		query := fmt.Sprintf("SELECT value FROM %s.kv_vise WHERE key = $1", pdb.schema)
		rs, err := tx.Query(ctx, query, lk.Translation)
		if err != nil {
			return nil, err
		}
		defer rs.Close()
		if rs.Next() {
			r := rs.RawValues()
			return r[0], nil
		}
	}

	tx, err := pdb.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT value FROM %s.kv_vise WHERE key = $1", pdb.schema)
	rs, err := tx.Query(ctx, query, lk.Translation)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	if !rs.Next() {
		return nil, db.NewErrNotFound(key)
	}
	r := rs.RawValues()
	return r[0], nil
}

// Close implements Db.
func(pdb *pgDb) Close() error {
	pdb.Close()
	return nil
}

// set up table
func(pdb *pgDb) prepare(ctx context.Context) error {
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
