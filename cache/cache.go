package cache

import (
	"fmt"
)

// Cache stores loaded content, enforcing size limits and keeping track of size usage.
// TODO: hide values from client, while allowing cbor serialization
type Cache struct {
	CacheSize uint32 // Total allowed cumulative size of values (not code) in cache
	CacheUseSize uint32 // Currently used bytes by all values (not code) in cache
	Cache []map[string]string // All loaded cache items
	Sizes map[string]uint16 // Size limits for all loaded symbols.
	LastValue string // last inserted value
	invalid bool
}

// NewCache creates a new ready-to-use cache object
func NewCache() *Cache {
	ca := &Cache{
		Cache: []map[string]string{make(map[string]string)},
		Sizes: make(map[string]uint16),
	}
	return ca
}

// Invalidate marks a cache as invalid.
//
// An invalid cache should not be persisted or propagated
func(ca *Cache) Invalidate() {
	ca.invalid = true
}

// Invalid returns true if cache is invalid.
//
// An invalid cache should not be persisted or propagated
func(ca *Cache) Invalid() bool {
	return ca.invalid
}

// WithCacheSize applies a cumulative cache size limitation for all cached items.
func(ca *Cache) WithCacheSize(cacheSize uint32) *Cache {
	ca.CacheSize = cacheSize
	return ca
}

// Add adds a cache value under a cache symbol key.
//
// Also stores the size limitation of for key for later updates.
//
// Fails if:
// - key already defined
// - value is longer than size limit
// - adding value exceeds cumulative cache capacity
func(ca *Cache) Add(key string, value string, sizeLimit uint16) error {
	if sizeLimit > 0 {
		l := uint16(len(value))
		if l > sizeLimit {
			return fmt.Errorf("value length %v exceeds value size limit %v", l, sizeLimit)
		}
	}
	checkFrame := ca.frameOf(key)
	if checkFrame > -1 {
		thisFrame := len(ca.Cache) - 1
		if checkFrame == thisFrame {
			Logg.Debugf("Ignoring load request on frame that has symbol already loaded")
			return nil
		}
		return fmt.Errorf("key %v already defined in frame %v, this is frame %v", key, checkFrame, thisFrame)
	}
	var sz uint32
	if len(value) > 0 {
		sz = ca.checkCapacity(value)
		if sz == 0 {
			return fmt.Errorf("Cache capacity exceeded %v of %v", ca.CacheUseSize + sz, ca.CacheSize)
		}
	}
	Logg.Infof("Cache add", "key", key, "size", sz, "limit", sizeLimit)
	Logg.Tracef("", "Cache add data", value)
	ca.Cache[len(ca.Cache)-1][key] = value
	ca.CacheUseSize += sz
	ca.Sizes[key] = sizeLimit
	ca.LastValue = value
	return nil
}

// ReservedSize returns the maximum byte size available for the given symbol.
func(ca *Cache) ReservedSize(key string) (uint16, error) {
	v, ok := ca.Sizes[key]
	if !ok {
		return 0, fmt.Errorf("unknown symbol: %s", key)
	}
	return v, nil
}

// Update sets a new value for an existing key.
//
// Uses the size limitation from when the key was added.
//
// Fails if:
// - key not defined
// - value is longer than size limit
// - replacing value exceeds cumulative cache capacity
func(ca *Cache) Update(key string, value string) error {
	sizeLimit := ca.Sizes[key]
	if ca.Sizes[key] > 0 {
		l := uint16(len(value))
		if l > sizeLimit {
			return fmt.Errorf("update value length %v exceeds value size limit %v", l, sizeLimit)
		}
	}
	checkFrame := ca.frameOf(key)
	if checkFrame == -1 {
		return fmt.Errorf("key %v not defined", key)
	}
	r := ca.Cache[checkFrame][key]
	l := uint32(len(r))
	ca.Cache[checkFrame][key] = ""
	ca.CacheUseSize -= l
	sz := ca.checkCapacity(value)
	if sz == 0 {
		baseUseSize := ca.CacheUseSize
		ca.Cache[checkFrame][key] = r
		ca.CacheUseSize += l
		return fmt.Errorf("Cache capacity exceeded %v of %v", baseUseSize + sz, ca.CacheSize)
	}
	ca.Cache[checkFrame][key] = value
	ca.CacheUseSize += uint32(len(value))
	return nil
}

// Get the content currently loaded for a single key, loaded at any level.
//
// Fails if key has not been loaded.
func(ca *Cache) Get(key string) (string, error) {
	i := ca.frameOf(key)
	if i == -1 {
		return "", fmt.Errorf("key '%s' not found in any frame", key)
	}
	r, ok := ca.Cache[i][key]
	if !ok {
		return "", fmt.Errorf("unknown key '%s'", key)
	}
	return r, nil
}

// Reset flushes all state contents below the top level.
func(ca *Cache) Reset() {
	if len(ca.Cache) == 0 {
		return
	}
	ca.Cache = ca.Cache[:1]
	ca.CacheUseSize = 0
	return
}

// Push adds a new level to the cache.
func (ca *Cache) Push() error {
	m := make(map[string]string)
	ca.Cache = append(ca.Cache, m)
	return nil
}

// Pop frees the cache of the current level and makes the previous level the current level.
//
// Fails if already on top level.
func (ca *Cache) Pop() error {
	l := len(ca.Cache)
	if l == 0 {
		return fmt.Errorf("already at top level")
	}
	l -= 1
	m := ca.Cache[l]
	for k, v := range m {
		sz := len(v)
		ca.CacheUseSize -= uint32(sz)
		Logg.Debugf("Cache free", "frame", l, "key", k, "size", sz)
	}
	ca.Cache = ca.Cache[:l]
	//ca.resetCurrent()
	return nil
}

// Check returns true if a key already exists in the cache.
func(ca *Cache) Check(key string) bool {
	return ca.frameOf(key) == -1
}

// Last returns the last inserted value
//
// The stored last inserter value will be reset to an empty string
// TODO: needs to be invalidated when out of scope
func(ca *Cache) Last() string {
	s := ca.LastValue
	ca.LastValue = ""
	return s
}

// bytes that will be added to cache use size for string
// returns 0 if capacity would be exceeded
func(ca *Cache) checkCapacity(v string) uint32 {
	sz := uint32(len(v))
	if ca.CacheSize == 0 {
		return sz
	}
	if ca.CacheUseSize + sz > ca.CacheSize {
		return 0	
	}
	return sz
}

// return 0-indexed frame number where key is defined. -1 if not defined
func(ca *Cache) frameOf(key string) int {
	for i, m := range ca.Cache {
		for k, _ := range m {
			if k == key {
				return i
			}
		}
	}
	return -1
}

func(ca *Cache) Levels() uint32 {
	return uint32(len(ca.Cache))
}

func(ca *Cache) Keys(level uint32) []string {
	var r []string
	for k := range ca.Cache[level] {
		r = append(r, k)
	}
	return r
}
