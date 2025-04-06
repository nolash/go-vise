package log

import (
	"context"
	"encoding/binary"
	"time"

	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/db"
)

var (
	logg logging.Logger = logging.NewVanilla().WithDomain("logdb")
)

type logDb struct {
	db.Db
	logDb db.Db
}

func NewLogDb(mainDb db.Db, db db.Db) db.Db {
	return &logDb{
		Db: mainDb,
		logDb: db,
	}
}

func (ldb *logDb) Start(ctx context.Context) error {
	err := ldb.Db.Start(ctx)
	if err != nil {
		return err	
	}
	err = ldb.logDb.Start(ctx)
	if err != nil {
		logg.DebugCtxf(ctx, "logdb start fail", "ctx", ctx, "err", err)
	}
	return nil
}

func (ldb *logDb) Stop(ctx context.Context) error {
	err := ldb.logDb.Stop(ctx)
	if err != nil {
		logg.DebugCtxf(ctx, "logdb stop fail", "ctx", ctx, "err", err)
	}
	return ldb.Db.Stop(ctx)
}

func (ldb *logDb) Connect(ctx context.Context, connStr string) error {
	err := ldb.Db.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	err = ldb.logDb.Connect(ctx, connStr)
	if err != nil {
		ldb.Db.Close(ctx)
	}
	return err
}

func (ldb *logDb) SetPrefix(pfx uint8) {
	ldb.Db.SetPrefix(pfx)	
	ldb.logDb.SetPrefix(pfx)	
}

func (ldb *logDb) SetLanguage(ln *lang.Language) {
	ldb.Db.SetLanguage(ln)
	ldb.logDb.SetLanguage(ln)
}

func (ldb *logDb) SetSession(sessionId string) {
	ldb.Db.SetSession(sessionId)	
	ldb.logDb.SetSession(sessionId)	
}

func (ldb *logDb) Base() *db.DbBase {
	return ldb.Db.Base()
}

func (ldb *logDb) toLogDbEntry(ctx context.Context, key []byte, val []byte) ([]byte, []byte) {
	var innerKey []byte
	var innerValKey []byte
	var innerValVal []byte
	var l []byte

	lk, err := ldb.Base().ToKey(ctx, key)
	if err != nil {
		return nil, nil
	}
	if lk.Translation == nil {
		innerValKey = lk.Default
	} else {
		innerValKey = lk.Translation
	}
	binary.PutUvarint(l, uint64(len(innerValKey)))
	innerValKey = append(l, innerValKey...)
	innerValKey = append(l, val...)

	t := time.Now().UnixNano()
	binary.BigEndian.PutUint64(innerKey, uint64(t))
	innerKey = ldb.Base().ToSessionKey(db.DATATYPE_UNKNOWN, innerKey)
	return innerKey, append(innerValKey, innerValVal...)
}

func (ldb *logDb) Put(ctx context.Context, key []byte, val []byte) error {
	err := ldb.Db.Put(ctx, key, val)
	if err != nil {
		return err
	}
	key, val = ldb.toLogDbEntry(ctx, key, val)
	if key == nil {
		logg.DebugCtxf(ctx, "logdb kv fail", "key", key, "err", err)
		return nil
	}
	err = ldb.logDb.Put(ctx, key, val)
	if err != nil {
		logg.DebugCtxf(ctx, "logdb put fail", "key", key, "err", err)
	}
	return nil
}
