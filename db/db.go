package db

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/lang"
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
	DATATYPE_USERSTART = 32
)

const (
	datatype_sessioned_threshold = DATATYPE_TEMPLATE
)

// Db abstracts all data storage and retrieval as a key-value store
type Db interface {
	// Connect prepares the storage backend for use. May panic or error if called more than once.
	Connect(ctx context.Context, connStr string) error
	// Close implements io.Closer. MUST be called before termination after a Connect().
	Close() error
	// Get retrieves the value belonging to a key. Errors if the key does not exist, or if the retrieval otherwise fails.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Put stores a value under a key. Any existing value will be replaced. Errors if the value could not be stored.
	Put(ctx context.Context, key []byte, val []byte) error
	// SetPrefix sets the storage context prefix to use for consecutive Get and Put operations.
	SetPrefix(pfx uint8)
	// SetSession sets the session context to use for consecutive Get and Put operations.
	// Session only affects the following datatypes:
	// * DATATYPE_STATE
	// * DATATYPE_USERSTART
	SetSession(sessionId string)
	SetLock(typ uint8, locked bool)
}

type lookupKey struct {
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
}

// ensures default locking of read-only entries
func(db *baseDb) defaultLock() {
	db.lock = DATATYPE_BIN | DATATYPE_MENU | DATATYPE_TEMPLATE | DATATYPE_STATICLOAD
}

// SetPrefix implements Db.
func(db *baseDb) SetPrefix(pfx uint8) {
	db.pfx = pfx
}

// SetSession implements Db.
func(db *baseDb) SetSession(sessionId string) {
	db.sid = append([]byte(sessionId), 0x2E)
}

// SetSafety disables modification of data that 
func(db *baseDb) SetLock(pfx uint8, lock bool) {
	if lock {
		db.lock	|= pfx
	} else {
		db.lock &= ^pfx
	}
}

func(db *baseDb) checkPut() bool {
	return db.pfx & db.lock == 0
}

func(db *baseDb) SetLanguage(ln *lang.Language) {
	db.lang = ln
}

// ToKey creates a DbKey within the current session context.
func(db *baseDb) ToKey(ctx context.Context, key []byte) (lookupKey, error) {
	var lk lookupKey
	var b []byte
	if db.pfx == DATATYPE_UNKNOWN {
		return lk, errors.New("datatype prefix cannot be UNKNOWN")
	}
	if (db.pfx > datatype_sessioned_threshold) {
		b = append(db.sid, key...)
	} else {
		b = key
	}
	lk.Default = ToDbKey(db.pfx, b, nil)
	if db.pfx & (DATATYPE_MENU | DATATYPE_TEMPLATE | DATATYPE_STATICLOAD) > 0 {
		ln, ok := ctx.Value("Language").(lang.Language)
		if ok {
			lk.Translation = ToDbKey(db.pfx, b, &ln)
		}
	}	
	return lk, nil
}
