package cache

import (
	"fmt"
	"log"
)

type Cache struct {
	CacheSize uint32 // Total allowed cumulative size of values (not code) in cache
	CacheUseSize uint32 // Currently used bytes by all values (not code) in cache
	Cache []map[string]string // All loaded cache items
	CacheMap map[string]string // Mapped
	menuSize uint16 // Max size of menu
	outputSize uint32 // Max size of output
	sizes map[string]uint16 // Size limits for all loaded symbols.
	sink *string
}

// NewCache creates a new ready-to-use cache object
func NewCache() *Cache {
	ca := &Cache{
		Cache: []map[string]string{make(map[string]string)},
		sizes: make(map[string]uint16),
	}
	ca.resetCurrent()
	return ca
}

// WithCacheSize applies a cumulative cache size limitation for all cached items.
func(ca *Cache) WithCacheSize(cacheSize uint32) *Cache {
	ca.CacheSize = cacheSize
	return ca
}

// WithCacheSize applies a cumulative cache size limitation for all cached items.
func(ca *Cache) WithOutputSize(outputSize uint32) *Cache {
	ca.outputSize = outputSize
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
		if checkFrame == len(ca.Cache) - 1 {
			log.Printf("Ignoring load request on frame that has symbol already loaded")
			return nil
		}
		return fmt.Errorf("key %v already defined in frame %v", key, checkFrame)
	}
	sz := ca.checkCapacity(value)
	if sz == 0 {
		return fmt.Errorf("Cache capacity exceeded %v of %v", ca.CacheUseSize + sz, ca.CacheSize)
	}
	log.Printf("add key %s value size %v limit %v", key, sz, sizeLimit)
	ca.Cache[len(ca.Cache)-1][key] = value
	ca.CacheUseSize += sz
	ca.sizes[key] = sizeLimit
	return nil
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
	sizeLimit := ca.sizes[key]
	if ca.sizes[key] > 0 {
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
	if ca.CacheMap[key] != "" {
		ca.CacheMap[key] = value
	}
	ca.CacheUseSize -= l
	sz := ca.checkCapacity(value)
	if sz == 0 {
		baseUseSize := ca.CacheUseSize
		ca.Cache[checkFrame][key] = r
		ca.CacheUseSize += l
		return fmt.Errorf("Cache capacity exceeded %v of %v", baseUseSize + sz, ca.CacheSize)
	}
	return nil
}

// Get returns the full key-value mapping for all mapped keys at the current cache level.
func(ca *Cache) Get() (map[string]string, error) {
	if len(ca.Cache) == 0 {
		return nil, fmt.Errorf("get at top frame")
	}
	return ca.Cache[len(ca.Cache)-1], nil
}

func(ca *Cache) Sizes() (map[string]uint16, error) {
	if len(ca.Cache) == 0 {
		return nil, fmt.Errorf("get at top frame")
	}
	sizes := make(map[string]uint16)
	var haveSink bool
	for k, _ := range ca.CacheMap {
		l, ok := ca.sizes[k]
		if !ok {
			panic(fmt.Sprintf("missing size for %v", k))
		}
		if l == 0 {
			if haveSink {
				panic(fmt.Sprintf("duplicate sink for %v", k))
			}
			haveSink = true
		}
		sizes[k] = l
	}
	return sizes, nil
}

// Map marks the given key for retrieval.
//
// After this, Val() will return the value for the key, and Size() will include the value size and limitations in its calculations.
//
// Only one symbol with no size limitation may be mapped at the current level.
func(ca *Cache) Map(key string) error {
	m, err := ca.Get()
	if err != nil {
		return err
	}
	l := ca.sizes[key]
	if l == 0 {
		if ca.sink != nil {
			return fmt.Errorf("sink already set to symbol '%v'", *ca.sink)
		}
		ca.sink = &key
	}
	ca.CacheMap[key] = m[key]
	return nil
}

// Fails if key is not mapped.
func(ca *Cache) Val(key string) (string, error) {
	r := ca.CacheMap[key]
	if len(r) == 0 {
		return "", fmt.Errorf("key %v not mapped", key)
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

// Size returns size used by values and menu, and remaining size available
func(ca *Cache) Usage() (uint32, uint32) {
	var l int
	var c uint16
	for k, v := range ca.CacheMap {
		l += len(v)
		c += ca.sizes[k]
	}
	r := uint32(l)
	r += uint32(ca.menuSize)
	return r, uint32(c)-r
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

// Push adds a new level to the cache.
func (ca *Cache) Push() error {
	m := make(map[string]string)
	ca.Cache = append(ca.Cache, m)
	ca.resetCurrent()
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
		log.Printf("free frame %v key %v value size %v", l, k, sz)
	}
	ca.Cache = ca.Cache[:l]
	ca.resetCurrent()
	return nil
}

// Check returns true if a key already exists in the cache.
func(ca *Cache) Check(key string) bool {
	return ca.frameOf(key) == -1
}

// flush relveant properties for level change
func(ca *Cache) resetCurrent() {
	ca.sink = nil
	ca.CacheMap = make(map[string]string)
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
