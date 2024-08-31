package cache

// Memory defines the interface for store of a symbol mapped content cache.
type Memory interface {
	// Add adds a cache value under a cache symbol key.
	//
	// Also stores the size limitation of for key for later updates.
	//
	// Must fail if:
	// 	* key already defined
	// 	* value is longer than size limit
	// 	* adding value exceeds cumulative cache capacity
	Add(key string, val string, sizeLimit uint16) error
	// Update sets a new value for an existing key.
	//
	// Uses the size limitation from when the key was added.
	//
	// Must fail if:
	// - key not defined
	// - value is longer than size limit
	// - replacing value exceeds cumulative cache capacity
	Update(key string, val string) error
	// ReservedSize returns the maximum byte size available for the given symbol.
	ReservedSize(key string) (uint16, error)
	// Get the content currently loaded for a single key, loaded at any level.
	//
	// Must fail if key has not been loaded.
	Get(key string) (string, error)
	// Push adds a new level to the cache.
	Push() error
	// Pop frees the cache of the current level and makes the previous level the current level.
	//
	// Fails if already on top level.
	Pop() error
	// Reset flushes all state contents below the top level.
	Reset()
	// Levels returns the current number of levels.
	Levels() uint32
	// Keys returns all storage keys for the given level.
	Keys(level uint32) []string
	// Last returns the last inserted value
	//
	// The stored last inserter value must be reset to an empty string
	Last() string
	// Invalidate marks a cache as invalid.
	//
	// An invalid cache should not be persisted or propagated
	Invalidate()
	// Invalid returns true if cache is invalid.
	//
	// An invalid cache should not be persisted or propagated
	Invalid() bool
}
