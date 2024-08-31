package db

import (
	"fmt"
)

// ErrNotFound is returned with a key was successfully queried, but did not match a stored key.
type ErrNotFound struct {
	k []byte
}

// NewErrNotFound creates a new ErrNotFound with the given storage key.
func NewErrNotFound(k []byte) error {
	return ErrNotFound{k}
}

// Error implements error.
func(e ErrNotFound) Error() string {
	return fmt.Sprintf("key not found: %x", e.k)
}
