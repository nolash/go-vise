package cache

import (
	"fmt"
)

// Cache stores loaded content, enforcing size limits and keeping track of size usage.
//
// TODO: hide values from client, while allowing cbor serialization
type Cache struct {
	// Total allowed cumulative size of values (not code) in cache
	CacheSize uint32
	// Currently used bytes by all values (not code) in cache
	CacheUseSize uint32
	// All loaded cache items
	Cache []map[string]string
	// Size limits for all loaded symbols.
	Sizes map[string]uint16
	// Last inserted value (regardless of scope)
	LastValue string 
	invalid bool
}

// NewCache creates a new ready-to-use Cache object
func NewCache() *Cache {
	ca := &Cache{
		Cache: []map[string]string{make(map[string]string)},
		Sizes: make(map[string]uint16),
	}
	return ca
}

// Invalidate implements the Memory interface.
func(ca *Cache) Invalidate() {
	ca.invalid = true
}

// Invalid implements the Memory interface.
func(ca *Cache) Invalid() bool {
	return ca.invalid
}

// WithCacheSize is a chainable method that applies a cumulative cache size limitation for all cached items.
func(ca *Cache) WithCacheSize(cacheSize uint32) *Cache {
	ca.CacheSize = cacheSize
	return ca
}

// Add implements the Memory interface.
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
			logg.Debugf("Ignoring load request on frame that has symbol already loaded")
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
	logg.Infof("Cache add", "key", key, "size", sz, "limit", sizeLimit)
	logg.Tracef("", "Cache add data", value)
	ca.Cache[len(ca.Cache)-1][key] = value
	ca.CacheUseSize += sz
	ca.Sizes[key] = sizeLimit
	ca.LastValue = value
	return nil
}

// ReservedSize implements the Memory interface.
func(ca *Cache) ReservedSize(key string) (uint16, error) {
	v, ok := ca.Sizes[key]
	if !ok {
		return 0, fmt.Errorf("unknown symbol: %s", key)
	}
	return v, nil
}

// Update implements the Memory interface.
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

// Get implements the Memory interface.
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

// Reset implements the Memory interface.
func(ca *Cache) Reset() {
	var v string
	if len(ca.Cache) == 0 {
		return
	}
	ca.Cache = ca.Cache[:1]
	ca.CacheUseSize = 0
	for _, v = range ca.Cache[0] {
		ca.CacheUseSize += uint32(len(v))
	}
	return
}

// Push implements the Memory interface.
func (ca *Cache) Push() error {
	m := make(map[string]string)
	ca.Cache = append(ca.Cache, m)
	return nil
}

// Pop implements the Memory interface.
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
		delete(ca.Sizes, k)
		logg.Debugf("Cache free", "frame", l, "key", k, "size", sz)
	}
	ca.Cache = ca.Cache[:l]
	if l == 0 {
		ca.Push()
	}
	return nil
}

// Check returns true if a key already exists in the cache.
func(ca *Cache) Check(key string) bool {
	return ca.frameOf(key) == -1
}

// Last implements the Memory interface.
//
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

// Levels implements the Memory interface.
func(ca *Cache) Levels() uint32 {
	return uint32(len(ca.Cache))
}

// Keys implements the Memory interface.
func(ca *Cache) Keys(level uint32) []string {
	var r []string
	for k := range ca.Cache[level] {
		r = append(r, k)
	}
	return r
}
