package postgres

import (
	"context"
	"errors"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"git.defalsify.org/vise.git/db"
)

var (
	defaultTxOptions pgx.TxOptions
)

type PgInterface interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Close()
}

// pgDb is a Postgres backend implementation of the Db interface.
type pgDb struct {
	*db.DbBase
	conn   PgInterface
	schema string
	prefix uint8
	prepd  bool
	it     pgx.Rows
	itBase []byte
	tx     pgx.Tx
	multi  bool
}

// NewpgDb creates a new Postgres backed Db implementation.
func NewPgDb() *pgDb {
	db := &pgDb{
		DbBase: db.NewDbBase(),
		schema: "public",
	}
	return db
}

// Base implements Db
func (pdb *pgDb) Base() *db.DbBase {
	return pdb.DbBase
}

// WithSchema sets the Postgres schema to use for the storage table.
func (pdb *pgDb) WithSchema(schema string) *pgDb {
	pdb.schema = schema
	return pdb
}

func (pdb *pgDb) WithConnection(pi PgInterface) *pgDb {
	pdb.conn = pi
	return pdb
}

// Connect implements Db.
func (pdb *pgDb) Connect(ctx context.Context, connStr string) error {
	if pdb.conn != nil {
		logg.WarnCtxf(ctx, "Pg already connected")
		return nil
	}
	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("connection to postgres could not be established: %w", err)
	}

	pdb.conn = conn
	pdb.DbBase.Connect(ctx, connStr)
	return pdb.ensureTable(ctx)
}

func (pdb *pgDb) Start(ctx context.Context) error {
	if pdb.tx != nil {
		return db.ErrTxExist
	}
	err := pdb.start(ctx)
	if err != nil {
		return err
	}
	pdb.multi = true
	return nil
}

func (pdb *pgDb) start(ctx context.Context) error {
	if pdb.tx != nil {
		return nil
	}
	tx, err := pdb.conn.BeginTx(ctx, defaultTxOptions)
	logg.TraceCtxf(ctx, "begin single tx", "err", err)
	if err != nil {
		return err
	}
	pdb.tx = tx
	return nil
}

func (pdb *pgDb) Stop(ctx context.Context) error {
	if !pdb.multi {
		return db.ErrSingleTx
	}
	return pdb.stop(ctx)
}

func (pdb *pgDb) stopSingle(ctx context.Context) error {
	if pdb.multi {
		return nil
	}
	err := pdb.tx.Commit(ctx)
	logg.TraceCtxf(ctx, "stop single tx", "err", err)
	pdb.tx = nil
	return err
}

func (pdb *pgDb) stop(ctx context.Context) error {
	if pdb.tx == nil {
		return db.ErrNoTx
	}
	err := pdb.tx.Commit(ctx)
	logg.TraceCtxf(ctx, "stop multi tx", "err", err)
	pdb.tx = nil
	return err
}

func (pdb *pgDb) Abort(ctx context.Context) {
	logg.InfoCtxf(ctx, "aborting tx", "tx", pdb.tx)
	pdb.tx.Rollback(ctx)
	pdb.tx = nil
}

// Put implements Db.
func (pdb *pgDb) Put(ctx context.Context, key []byte, val []byte) error {
	if !pdb.CheckPut() {
		return errors.New("unsafe put and safety set")
	}

	lk, err := pdb.ToKey(ctx, key)
	if err != nil {
		return err
	}

	err = pdb.start(ctx)
	if err != nil {
		return err
	}
	logg.TraceCtxf(ctx, "put", "key", key, "val", val)
	query := fmt.Sprintf("INSERT INTO %s.kv_vise (key, value, updated) VALUES ($1, $2, 'now') ON CONFLICT(key) DO UPDATE SET value = $2, updated = 'now';", pdb.schema)
	actualKey := lk.Default
	if lk.Translation != nil {
		actualKey = lk.Translation
	}

	_, err = pdb.tx.Exec(ctx, query, actualKey, val)
	if err != nil {
		return err
	}

	return pdb.stopSingle(ctx)
}

// Get implements Db.
func (pdb *pgDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	var rr []byte
	lk, err := pdb.ToKey(ctx, key)
	if err != nil {
		return nil, err
	}

	err = pdb.start(ctx)
	if err != nil {
		return nil, err
	}
	logg.TraceCtxf(ctx, "get", "key", key)

	if lk.Translation != nil {
		query := fmt.Sprintf("SELECT value FROM %s.kv_vise WHERE key = $1", pdb.schema)
		rs, err := pdb.tx.Query(ctx, query, lk.Translation)
		if err != nil {
			pdb.Abort(ctx)
			return nil, err
		}

		if rs.Next() {
			err = rs.Scan(&rr)
			if err != nil {
				pdb.Abort(ctx)
				return nil, err
			}

			rs.Close()
			err = pdb.stopSingle(ctx)
			return rr, err
		}
	}

	query := fmt.Sprintf("SELECT value FROM %s.kv_vise WHERE key = $1", pdb.schema)
	rs, err := pdb.tx.Query(ctx, query, lk.Default)
	if err != nil {
		pdb.Abort(ctx)
		return nil, err
	}

	if !rs.Next() {
		rs.Close()
		pdb.Abort(ctx)
		return nil, db.NewErrNotFound(key)
	}

	err = rs.Scan(&rr)
	if err != nil {
		rs.Close()
		pdb.Abort(ctx)
		return nil, err
	}
	rs.Close()
	err = pdb.stopSingle(ctx)
	return rr, err
}

// Close implements Db.
func (pdb *pgDb) Close(ctx context.Context) error {
	err := pdb.Stop(ctx)
	if err == db.ErrNoTx {
		err = nil
	}
	pdb.conn.Close()
	return err
}

// set up table
func (pdb *pgDb) ensureTable(ctx context.Context) error {
	if pdb.prepd {
		logg.WarnCtxf(ctx, "ensureTable called more than once")
		return nil
	}
	tx, err := pdb.conn.BeginTx(ctx, defaultTxOptions)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kv_vise (
		id SERIAL NOT NULL,
		key BYTEA NOT NULL UNIQUE,
		value BYTEA NOT NULL,
		updated TIMESTAMP NOT NULL
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
	pdb.prepd = true
	return nil
}
