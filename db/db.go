package db

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"git.defalsify.org/vise.git/lang"
)

const (
	safeLock = DATATYPE_BIN | DATATYPE_MENU | DATATYPE_TEMPLATE | DATATYPE_STATICLOAD
)

const (
	// Invalid datatype, must raise error if attempted used.
	DATATYPE_UNKNOWN = 0
	// Bytecode
	DATATYPE_BIN = 1
	// Menu symbol
	DATATYPE_MENU = 2
	// Template symbol
	DATATYPE_TEMPLATE = 4
	// Static LOAD symbols
	DATATYPE_STATICLOAD = 8
	// State and cache from persister
	DATATYPE_STATE = 16
	// Application data
	DATATYPE_USERDATA = 32
)

const (
	datatype_sessioned_threshold = DATATYPE_STATICLOAD
)

// Db abstracts all data storage and retrieval as a key-value store
type Db interface {
	// Connect prepares the storage backend for use.
	//
	// If called more than once, consecutive calls should be ignored.
	Connect(ctx context.Context, connStr string) error
	// MUST be called before termination after a Connect().
	Close(context.Context) error
	// Get retrieves the value belonging to a key.
	//
	// Errors if the key does not exist, or if the retrieval otherwise fails.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Put stores a value under a key.
	//
	// Any existing value will be replaced.
	//
	// Errors if the value could not be stored.
	Put(ctx context.Context, key []byte, val []byte) error
	// SetPrefix sets the storage context prefix to use for consecutive Get and Put operations.
	SetPrefix(pfx uint8)
	// SetSession sets the session context to use for consecutive Get and Put operations.
	//
	// Session only affects the following datatypes:
	// * DATATYPE_STATE
	// * DATATYPE_USERSTART
	SetSession(sessionId string)
	// SetLock disables modification of data that is readonly in the vm context.
	//
	// If called with typ value 0, it will permanently lock all readonly members.
	SetLock(typ uint8, locked bool) error
	// Safe returns true if db is safe for use with a vm.
	Safe() bool
	// SetLanguage sets the language context to use on consecutive gets or puts
	//
	// Language only affects the following datatypes:
	// * DATATYPE_MENU
	// * DATATYPE_TEMPLATE
	// * DATATYPE_STATICLOAD
	SetLanguage(*lang.Language)
	// Prefix returns the current active datatype prefix
	Prefix() uint8
	// Dump generates an iterable dump of all keys matching the given byte prefix.
	Dump(context.Context, []byte) (*Dumper, error)
	// DecodeKey decodes the database specific key used for internal storage to the original key given by the caller.
	DecodeKey(ctx context.Context, key []byte) ([]byte, error)
	// Start creates a new database transaction. Only relevant for transactional databases.
	Start(context.Context) error
	// Stop completes a database transaction. Only relevant for transactional databases.
	Stop(context.Context) error
	// Abort cancels a database transaction. Only relevant for transactional databases.
	Abort(context.Context)
	// Connection returns the complete connection string used by the database implementation to connect.
	Connection() string
	// Base returns the underlying DbBase
	Base() *DbBase
}

// LookupKey encapsulates two keys for a database entry; one for the default language, the other for the language in the context at which the LookupKey was generated.
type LookupKey struct {
	Default     []byte
	Translation []byte
}

// ToDbKey generates a key to use Db to store a value for a particular context.
//
// If language is nil, then default language storage context will be used.
//
// If language is not nil, and the context does not support language, the language value will silently will be ignored.
func ToDbKey(typ uint8, b []byte, l *lang.Language) []byte {
	k := []byte{typ}
	if l != nil && l.Code != "" && typ&(DATATYPE_MENU|DATATYPE_TEMPLATE|DATATYPE_STATICLOAD) > 0 {
		b = append(b, []byte("_"+l.Code)...)
		//s += "_" + l.Code
	}
	return append(k, b...)
}

// FromDbKey parses the storage key as used in the database implementation to the original key bytes given by the client code.
func FromDbKey(b []byte) ([]byte, error) {
	if len(b) < 2 {
		return nil, fmt.Errorf("invalid db key")
	}
	typ := b[0]
	b = b[1:]
	if typ&(DATATYPE_MENU|DATATYPE_TEMPLATE|DATATYPE_STATICLOAD) > 0 {
		if len(b) > 6 {
			if b[len(b)-4] == '_' {
				b = b[:len(b)-4]
			}
		}
	}
	return b, nil
}

// baseDb is a base class for all Db implementations.
type baseDb struct {
	pfx     uint8
	sid     []byte
	lock    uint8
	lang    *lang.Language
	seal    bool
	connStr string
	known bool
	logDb Db
}

// DbBase is a base class that must be extended by all db.Db implementers.
//
// It must be created with NewDbBase()
type DbBase struct {
	*baseDb
}

// NewDbBase instantiates a new DbBase.
func NewDbBase() *DbBase {
	db := &DbBase{
		baseDb: &baseDb{
			known: true,
		},
	}
	db.baseDb.defaultLock()
	return db
}

// AllowUnknownPrefix disables the error generated when the DATATYPE_UNKNOWN prefix is used for storage.
func (db *baseDb) AllowUnknownPrefix() bool {
	known := db.known
	if !known {
		return false
	}
	db.known = false
	return true
}

// ensures default locking of read-only entries
func (db *baseDb) defaultLock() {
	db.lock |= safeLock
}

// Safe returns true if the database has been set up to protect read-only data.
func (bd *DbBase) Safe() bool {
	return bd.baseDb.lock&safeLock == safeLock
}

// Prefix returns the current prefix (last set by SetPrefix)
func (bd *DbBase) Prefix() uint8 {
	return bd.baseDb.pfx
}

// SetPrefix implements the Db interface.
func (bd *DbBase) SetPrefix(pfx uint8) {
	bd.baseDb.pfx = pfx
}

// SetLanguage implements the Db interface.
func (bd *DbBase) SetLanguage(ln *lang.Language) {
	bd.baseDb.lang = ln
}

// SetSession implements the Db interface.
func (bd *DbBase) SetSession(sessionId string) {
	if sessionId == "" {
		bd.baseDb.sid = []byte{}
	} else {
		bd.baseDb.sid = append([]byte(sessionId), 0x2E)
	}
}

// SetLock implements the Db interface.
func (bd *DbBase) SetLock(pfx uint8, lock bool) error {
	if bd.baseDb.seal {
		return errors.New("SetLock on sealed db")
	}
	if pfx == 0 {
		bd.baseDb.defaultLock()
		bd.baseDb.seal = true
		return nil
	}
	if lock {
		bd.baseDb.lock |= pfx
	} else {
		bd.baseDb.lock &= ^pfx
	}
	return nil
}

// CheckPut returns true if the current selected data type can be written to.
func (bd *DbBase) CheckPut() bool {
	return bd.baseDb.pfx&bd.baseDb.lock == 0
}

// ToSessionKey applies the currently set session id to the key.
//
// If the key in pfx does not use session, the key is returned unchanged.
func (bd *DbBase) ToSessionKey(pfx uint8, key []byte) []byte {
	var b []byte
	if pfx > datatype_sessioned_threshold || pfx == DATATYPE_UNKNOWN {
		b = append([]byte(bd.sid), key...)
	} else {
		b = key
	}
	return b
}

// FromSessionKey reverses the effect of ToSessionKey.
func (bd *DbBase) FromSessionKey(key []byte) ([]byte, error) {
	if len(bd.baseDb.sid) == 0 {
		return key, nil
	}
	if !bytes.HasPrefix(key, bd.baseDb.sid) {
		return nil, fmt.Errorf("session id prefix %s does not match key %x", string(bd.baseDb.sid), key)
	}
	return bytes.TrimPrefix(key, bd.baseDb.sid), nil
}

// ToKey creates a DbKey within the current session context.
//
// TODO: hard to read, clean up
func (bd *DbBase) ToKey(ctx context.Context, key []byte) (LookupKey, error) {
	var ln *lang.Language
	var lk LookupKey
	//var b []byte
	db := bd.baseDb
	if db.known && db.pfx == DATATYPE_UNKNOWN {
		return lk, errors.New("datatype prefix cannot be UNKNOWN")
	}
	//b := ToSessionKey(db.pfx, db.sid, key)
	b := bd.ToSessionKey(db.pfx, key)
	lk.Default = ToDbKey(db.pfx, b, nil)
	if db.pfx&(DATATYPE_MENU|DATATYPE_TEMPLATE|DATATYPE_STATICLOAD) > 0 {
		if db.lang != nil {
			ln = db.lang
		} else {
			lo, ok := ctx.Value("Language").(lang.Language)
			if ok {
				ln = &lo
			}
		}
		logg.TraceCtxf(ctx, "language using", "ln", ln)
		if ln != nil {
			lk.Translation = ToDbKey(db.pfx, b, ln)
		}
	}
	logg.TraceCtxf(ctx, "made db lookup key", "lk", lk.Default, "pfx", db.pfx)
	return lk, nil
}

// DecodeKey implements Db.
func (bd *DbBase) DecodeKey(ctx context.Context, key []byte) ([]byte, error) {
	var err error
	oldKey := key
	key, err = FromDbKey(key)
	if err != nil {
		return []byte{}, err
	}
	key, err = bd.FromSessionKey(key)
	if err != nil {
		return []byte{}, err
	}
	logg.DebugCtxf(ctx, "decoded key", "key", key, "fromkey", oldKey)
	return key, nil
}

// Start implements Db.
func (bd *DbBase) Start(ctx context.Context) error {
	return nil
}

// Stop implements Db.
func (bd *DbBase) Stop(ctx context.Context) error {
	return nil
}

// Abort implements Db.
func (bd *DbBase) Abort(ctx context.Context) {
}

// Connect implements Db.
func (bd *DbBase) Connect(ctx context.Context, connStr string) error {
	bd.connStr = connStr
	return nil
}

// Connection implements Db.
func (bd *DbBase) Connection() string {
	return bd.connStr
}
