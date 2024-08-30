package db

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/lang"
)

const (
	DATATYPE_UNKNOWN = 0
	DATATYPE_BIN = 1
	DATATYPE_MENU = 2
	DATATYPE_TEMPLATE = 4
	DATATYPE_STATE = 8
	DATATYPE_USERSTART = 16
)

const (
	datatype_sessioned_threshold = DATATYPE_TEMPLATE
)

// Db abstracts all data storage and retrieval as a key-value store
type Db interface {
	// Connect prepares the storage backend for use
	Connect(ctx context.Context, connStr string) error
	// Close implements io.Closer
	Close() error
	// Get retrieves the value belonging to a key. Errors if the key does not exist, or if the retrieval otherwise fails.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Put stores a value under a key. Any existing value will be replaced. Errors if the value could not be stored.
	Put(ctx context.Context, key []byte, val []byte) error
	// SetPrefix sets the storage context prefix to use for consecutive Get and Put operations.
	SetPrefix(pfx uint8)
	// SetSession sets the session context to use for consecutive Get and Put operations.
	SetSession(sessionId string)
}

// ToDbKey generates a key to use Db to store a value for a particular context.
//
// If language is nil, then default language storage context will be used.
//
// If language is not nil, and the context does not support language, the language value will silently will be ignored.
func ToDbKey(typ uint8, b []byte, l *lang.Language) []byte {
	k := []byte{typ}
	if l != nil && l.Code != "" {
		k = append(k, []byte("_" + l.Code)...)
		//s += "_" + l.Code
	}
	return append(k, b...)
}

// BaseDb is a base class for all Db implementations.
type BaseDb struct {
	pfx uint8
	sid []byte
}

// SetPrefix implements Db.
func(db *BaseDb) SetPrefix(pfx uint8) {
	db.pfx = pfx
}

// SetSession implements Db.
func(db *BaseDb) SetSession(sessionId string) {
	db.sid = append([]byte(sessionId), 0x2E)
}

// ToKey creates a DbKey within the current session context.
func(db *BaseDb) ToKey(key []byte) ([]byte, error) {
	var b []byte
	if db.pfx == DATATYPE_UNKNOWN {
		return nil, errors.New("datatype prefix cannot be UNKNOWN")
	}
	if (db.pfx > datatype_sessioned_threshold) {
		b = append(db.sid, key...)
	} else {
		b = key
	}
	return ToDbKey(db.pfx, b, nil), nil
}
