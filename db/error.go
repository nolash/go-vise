package db

import (
	"errors"
	"fmt"
	"strings"
)

const (
	notFoundPrefix = "key not found: "
)

var (
	ErrTxExist = errors.New("tx already exists")
	ErrNoTx = errors.New("tx does not exist")
	ErrSingleTx = errors.New("not a multi-instruction tx")
)

// ErrNotFound is returned with a key was successfully queried, but did not match a stored key.
type ErrNotFound struct {
	k []byte
}

// NewErrNotFound creates a new ErrNotFound with the given storage key.
func NewErrNotFound(k []byte) error {
	return ErrNotFound{k}
}

// Error implements Error.
func(e ErrNotFound) Error() string {
	return fmt.Sprintf("%s%x", notFoundPrefix, e.k)
}

func (e ErrNotFound) Is(err error) bool {
	return strings.Contains(err.Error(), notFoundPrefix)
}

func IsNotFound(err error) bool {
	target := ErrNotFound{}
	return target.Is(err)
}
