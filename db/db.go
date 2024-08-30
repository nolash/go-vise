package db

import (
	"context"

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
