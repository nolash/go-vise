package db

import (
	"git.defalsify.org/vise.git/lang"
)

const (
	DATATYPE_UNKNOWN = iota
	DATATYPE_BIN
	DATATYPE_TEMPLATE
	DATATYPE_STATE
)

type Db interface {
	Connect(connStr string) error
	Get(key []byte) ([]byte, error)
	Put(key []byte, val []byte) error
}

func ToDbKey(typ uint8, s string, l *lang.Language) []byte {
	k := []byte{typ}
	if l != nil && l.Code != "" {
		s += "_" + l.Code
	}
	return append(k, []byte(s)...)
}
