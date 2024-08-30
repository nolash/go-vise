package db

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/lang"
)

const (
	DATATYPE_UNKNOWN = iota
	DATATYPE_BIN
	DATATYPE_TEMPLATE
	DATATYPE_STATE
	DATATYPE_USERSTART
)

type Db interface {
	Connect(ctx context.Context, connStr string) error
	Close() error
	Get(ctx context.Context, sessionId string, key []byte) ([]byte, error)
	Put(ctx context.Context, sessionId string, key []byte, val []byte) error
}

func ToDbKey(typ uint8, b []byte, l *lang.Language) []byte {
	k := []byte{typ}
	if l != nil && l.Code != "" {
		k = append(k, []byte("_" + l.Code)...)
		//s += "_" + l.Code
	}
	return append(k, b...)
}

type BaseDb struct {
	pfx uint8
}

func(db *BaseDb) SetPrefix(pfx uint8) {
	db.pfx = pfx
}

func(db *BaseDb) ToKey(sessionId string, key []byte) ([]byte, error) {
	if db.pfx == DATATYPE_UNKNOWN {
		return nil, errors.New("datatype prefix must be set explicitly")
	}
	b := append([]byte(sessionId), 0x2E)
	b = append(b, key...)
	return ToDbKey(db.pfx, b, nil), nil
}
