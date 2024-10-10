package db

import (
	"context"
	"errors"

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
	// Close implements io.Closer.
	//
	// MUST be called before termination after a Connect().
	Close() error
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
}

type LookupKey struct {
	Default []byte
	Translation []byte
}

// ToDbKey generates a key to use Db to store a value for a particular context.
//
// If language is nil, then default language storage context will be used.
//
// If language is not nil, and the context does not support language, the language value will silently will be ignored.
func ToDbKey(typ uint8, b []byte, l *lang.Language) []byte {
	k := []byte{typ}
	if l != nil && l.Code != "" && typ & (DATATYPE_MENU | DATATYPE_TEMPLATE | DATATYPE_STATICLOAD) > 0 {
		b = append(b, []byte("_" + l.Code)...)
		//s += "_" + l.Code
	}
	return append(k, b...)
}

// baseDb is a base class for all Db implementations.
type baseDb struct {
	pfx uint8
	sid []byte
	lock uint8
	lang *lang.Language
	seal bool
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
		baseDb: &baseDb{},
	}
	db.baseDb.defaultLock()
	return db
}

// ensures default locking of read-only entries
func(db *baseDb) defaultLock() {
	db.lock |= safeLock
}

func(bd *DbBase) Safe() bool {
	return bd.baseDb.lock & safeLock == safeLock
}

func(bd *DbBase) Prefix() uint8 {
	return bd.baseDb.pfx
}

// SetPrefix implements the Db interface.
func(bd *DbBase) SetPrefix(pfx uint8) {
	bd.baseDb.pfx = pfx
}

// SetLanguage implements the Db interface.
func(bd *DbBase) SetLanguage(ln *lang.Language) {
	bd.baseDb.lang = ln
}
// SetSession implements the Db interface.
func(bd *DbBase) SetSession(sessionId string) {
	bd.baseDb.sid = append([]byte(sessionId), 0x2E)
}

// SetLock implements the Db interface.
func(bd *DbBase) SetLock(pfx uint8, lock bool) error {
	if bd.baseDb.seal {
		return errors.New("SetLock on sealed db")
	}
	if pfx == 0 {
		bd.baseDb.defaultLock()
		bd.baseDb.seal = true
		return nil
	}
	if lock {
		bd.baseDb.lock	|= pfx
	} else {
		bd.baseDb.lock &= ^pfx
	}
	return nil
}

// CheckPut returns true if the current selected data type can be written to.
func(bd *DbBase) CheckPut() bool {
	return bd.baseDb.pfx & bd.baseDb.lock == 0
}

func ToSessionKey(pfx uint8, sessionId []byte, key []byte) []byte {
	var b []byte
	if (pfx > datatype_sessioned_threshold) {
		b = append([]byte(sessionId), key...)
	} else {
		b = key
	}
	return b
}

// ToKey creates a DbKey within the current session context.
//
// TODO: hard to read, clean up
func(bd *DbBase) ToKey(ctx context.Context, key []byte) (LookupKey, error) {
	var ln *lang.Language
	var lk LookupKey
	//var b []byte
	db := bd.baseDb
	if db.pfx == DATATYPE_UNKNOWN {
		return lk, errors.New("datatype prefix cannot be UNKNOWN")
	}
	b := ToSessionKey(db.pfx, db.sid, key)
	lk.Default = ToDbKey(db.pfx, b, nil)
	if db.pfx & (DATATYPE_MENU | DATATYPE_TEMPLATE | DATATYPE_STATICLOAD) > 0 {
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
	return lk, nil
}
