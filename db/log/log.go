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

// TODO: add a formatted dumper for the logdb
type logDb struct {
	db.Db
	logDb db.Db
}

type LogEntry struct {
	Key []byte
	Val []byte
	SessionId string
	When time.Time
	Pfx uint8
}

// NewLogDb creates a wrapper for the main Db in the first argument, which write an entry for every Put to the second database.
//
// All interface methods operate like normal on the main Db.
//
// Errors writing to the log database are ignored (but logged).
// 
// The Put is recorded in the second database under a chronologically sorted session key:
//
// `db.DATATYPE_UNKNOWN | sessionId | "_" | Big-endian uint64 representation of nanoseconds of time of put`
// 
// The value is stored as:
//
// `varint(length(key)) | key | value`
func NewLogDb(mainDb db.Db, subDb db.Db) *logDb {
	subDb.Base().AllowUnknownPrefix()
	return &logDb{
		Db: mainDb,
		logDb: subDb,
	}
}

// Start implements Db
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

// Stop implements Db
func (ldb *logDb) Stop(ctx context.Context) error {
	err := ldb.logDb.Stop(ctx)
	if err != nil {
		logg.DebugCtxf(ctx, "logdb stop fail", "ctx", ctx, "err", err)
	}
	return ldb.Db.Stop(ctx)
}

// Connect implements Db.
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

// SetLanguage implements Db.
func (ldb *logDb) SetLanguage(ln *lang.Language) {
	ldb.Db.SetLanguage(ln)
	ldb.logDb.SetLanguage(ln)
}

// SetSession implements Db.
func (ldb *logDb) SetSession(sessionId string) {
	ldb.Db.SetSession(sessionId)	
	ldb.logDb.SetSession(sessionId)	
}

// Base implement Db.
func (ldb *logDb) Base() *db.DbBase {
	return ldb.Db.Base()
}

// create the chronological logentry key to store the put under.
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

	l = make([]byte, 8)
	c := binary.PutUvarint(l, uint64(len(innerValKey)))
	innerValKey = append(l[:c], innerValKey...)
	innerValKey = append(innerValKey, val...)

	innerKey = make([]byte, 8)
	t := time.Now().UnixNano()
	binary.BigEndian.PutUint64(innerKey, uint64(t))
	return innerKey, append(innerValKey, innerValVal...)
}

// ToLogDbEntry decodes a logdb entry to a structure containing the relevant metadata aswell as the original key and value pass by the client.
func (ldb *logDb) ToLogDbEntry(ctx context.Context, key []byte, val []byte) (LogEntry, error) {
	var err error

	key = key[1:]
	tb := key[len(key)-8:]
	nsecs := binary.BigEndian.Uint64(tb[:8])
	nsecPart := int64(nsecs % 1000000000)
	secPart := int64(nsecs / 1000000000)
	t := time.Unix(secPart, nsecPart)

	sessionId := key[:len(key)-8]

	l, c := binary.Uvarint(val)
	lk := val[c:uint64(c)+l]
	v := val[uint64(c)+l:]
	pfx := lk[0]
	k, err := ldb.Base().DecodeKey(ctx, lk)
	if err != nil {
		return LogEntry{}, err
	}
	sessionId = lk[1:len(lk)-len(k)-1]
	return LogEntry{
		Key: k,
		Val: v,
		SessionId: string(sessionId),
		When: t,
		Pfx: pfx,
	}, nil
}

// Put implements Db.
func (ldb *logDb) Put(ctx context.Context, key []byte, val []byte) error {
	ldb.logDb.SetPrefix(db.DATATYPE_UNKNOWN)
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
