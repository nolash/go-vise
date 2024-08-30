package db

import (
	"fmt"
)

type ErrNotFound struct {
	k []byte
}

func NewErrNotFound(k []byte) error {
	return ErrNotFound{k}
}

func(e ErrNotFound) Error() string {
	return fmt.Sprintf("key not found: %x", e.k)
}
