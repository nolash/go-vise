package persist

import (
	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/render"
	"git.defalsify.org/vise.git/state"
)

// Persister interface defines the methods needed for a component that can store the execution state to a storage location.
type Persister interface {
	Serialize() ([]byte, error) // Output serializes representation of the state.
	Deserialize(b []byte) error // Restore state from a serialized state.
	Save(key string, renderer render.Renderer) error // Serialize and commit the state representation to persisted storage.
	Load(key string) error // Load the state representation from persisted storage and Deserialize.
	GetState() *state.State // Get the currently loaded State object.
	GetMemory() cache.Memory // Get the currently loaded Cache object.
	GetKeys() []string // Get all mapped keys for renderer.
}

