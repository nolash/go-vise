package persist

import (
	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/state"
)

type Persister interface {
	Serialize() ([]byte, error)
	Deserialize(b []byte) error
	Save(key string) error
	Load(key string) error
	GetState() *state.State
	GetMemory() cache.Memory
}

