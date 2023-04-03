package state

import (
	"fmt"
	"log"
)

// State holds the command stack, error condition of a unique execution session.
//
// It also holds cached values for all results of executed symbols.
//
// Cached values are linked to the command stack level it which they were loaded. When they go out of scope they are freed.
//
// Values must be mapped to a level in order to be available for retrieval and count towards size
//
// It can hold a single argument, which is freed once it is read
//
// Symbols are loaded with individual size limitations. The limitations apply if a load symbol is updated. Symbols may be added with a 0-value for limits, called a "sink." If mapped, the sink will consume all net remaining size allowance unused by other symbols. Only one sink may be mapped per level.
//
// Symbol keys do not count towards cache size limitations.
//
// 8 first flags are reserved.
//
// TODO factor out cache
type State struct {
	Flags []byte // Error state
	CacheSize uint32 // Total allowed cumulative size of values (not code) in cache
	CacheUseSize uint32 // Currently used bytes by all values (not code) in cache
	Cache []map[string]string // All loaded cache items
	CacheMap map[string]string // Mapped
	input []byte // Last input
	code []byte // Pending bytecode to execute
	execPath []string // Command symbols stack
	arg *string // Optional argument. Nil if not set.
	sizes map[string]uint16 // Size limits for all loaded symbols.
	sink *string // Sink symbol set for level
	bitSize uint32 // size of (32-bit capacity) bit flag byte array
	//sizeIdx uint16
}

func toByteSize(bitSize uint32) uint8 {
	if bitSize == 0 {
		return 0
	}
	n := bitSize % 8
	if n > 0 {
		bitSize += (8 - n)
	}
	return uint8(bitSize / 8)
}

// Retrieve the state of a state flag
func getFlag(bitIndex uint32, bitField []byte) bool {
	byteIndex := bitIndex / 8
	localBitIndex := bitIndex % 8
	b := bitField[byteIndex]
	return (b & (1 << localBitIndex)) > 0
}

// NewState creates a new State object with bitSize number of error condition states in ADDITION to the 8 builtin flags.
func NewState(bitSize uint32) State {
	st := State{
		CacheSize: 0,
		CacheUseSize: 0,
		bitSize: bitSize + 8,
	}
	byteSize := toByteSize(bitSize + 8)
	if byteSize > 0 {
		st.Flags = make([]byte, byteSize) 
	} else {
		st.Flags = []byte{}
	}
	st.Down("")
	return st
}

// SetFlag sets the flag at the given bit field index
//
// Returns true if bit state was changed.
//
// Fails if bitindex is out of range.
func(st *State) SetFlag(bitIndex uint32) (bool, error) {
	if bitIndex + 1 > st.bitSize {
		return false, fmt.Errorf("bit index %v is out of range of bitfield size %v", bitIndex, st.bitSize)
	}
	r := getFlag(bitIndex, st.Flags)
	if r {
		return false, nil
	}
	byteIndex := bitIndex / 8
	localBitIndex := bitIndex % 8
	b := st.Flags[byteIndex] 
	st.Flags[byteIndex] = b | (1 << localBitIndex)
	return true, nil
}


// ResetFlag resets the flag at the given bit field index.
//
// Returns true if bit state was changed.
//
// Fails if bitindex is out of range.
func(st *State) ResetFlag(bitIndex uint32) (bool, error) {
	if bitIndex + 1 > st.bitSize {
		return false, fmt.Errorf("bit index %v is out of range of bitfield size %v", bitIndex, st.bitSize)
	}
	r := getFlag(bitIndex, st.Flags)
	if !r {
		return false, nil
	}
	byteIndex := bitIndex / 8
	localBitIndex := bitIndex % 8
	b := st.Flags[byteIndex] 
	st.Flags[byteIndex] = b & (^(1 << localBitIndex))
	return true, nil
}

// GetFlag returns the state of the flag at the given bit field index.
//
// Fails if bit field index is out of range.
func(st *State) GetFlag(bitIndex uint32) (bool, error) {
	if bitIndex + 1 > st.bitSize {
		return false, fmt.Errorf("bit index %v is out of range of bitfield size %v", bitIndex, st.bitSize)
	}
	return getFlag(bitIndex, st.Flags), nil
}

// FlagBitSize reports the amount of bits available in the bit field index.
func(st *State) FlagBitSize() uint32 {
	return st.bitSize
}

// FlagBitSize reports the amount of bits available in the bit field index.
func(st *State) FlagByteSize() uint8 {
	return uint8(len(st.Flags))
}

// GetIndex scans a byte slice in same order as in storage, and returns the index of the first set bit.
//
// If the given byte slice is too small for the bit field bitsize, the check will terminate at end-of-data without error.
func(st *State) GetIndex(flags []byte) bool {
	var globalIndex uint32
	if st.bitSize == 0 {
		return false
	}
	if len(flags) == 0 {
		return false
	}
	var byteIndex uint8
	var localIndex uint8
	l := uint8(len(flags))
	var i uint32
	for i = 0; i < st.bitSize; i++ {
		testVal := flags[byteIndex] & (1 << localIndex)
		if (testVal & st.Flags[byteIndex]) > 0 {
			return true
		}
		globalIndex += 1
		if globalIndex % 8 == 0 {
			byteIndex += 1
			localIndex = 0
			if byteIndex > (l - 1) {
				return false				
			}
		} else {
			localIndex += 1
		}
	}
	return false
}

// WithCacheSize applies a cumulative cache size limitation for all cached items.
func(st State) WithCacheSize(cacheSize uint32) State {
	st.CacheSize = cacheSize
	return st
}

// Where returns the current active rendering symbol.
func(st State) Where() string {
	if len(st.execPath) == 0 {
		return ""
	}
	l := len(st.execPath)
	return st.execPath[l-1]
}

// Down adds the given symbol to the command stack.
//
// Clears mapping and sink.
func(st *State) Down(input string) {
	m := make(map[string]string)
	st.Cache = append(st.Cache, m)
	st.sizes = make(map[string]uint16)
	st.execPath = append(st.execPath, input)
	st.resetCurrent()
}


// Up removes the latest symbol to the command stack, and make the previous symbol current.
//
// Frees all symbols and associated values loaded at the previous stack level. Cache capacity is increased by the corresponding amount.
//
// Clears mapping and sink.
//
// Fails if called at top frame.
func(st *State) Up() error {
	l := len(st.Cache)
	if l == 0 {
		return fmt.Errorf("exit called beyond top frame")
	}
	l -= 1
	m := st.Cache[l]
	for k, v := range m {
		sz := len(v)
		st.CacheUseSize -= uint32(sz)
		log.Printf("free frame %v key %v value size %v", l, k, sz)
	}
	st.Cache = st.Cache[:l]
	st.execPath = st.execPath[:l]
	st.resetCurrent()
	return nil
}

// Add adds a cache value under a cache symbol key.
//
// Also stores the size limitation of for key for later updates.
//
// Fails if:
// - key already defined
// - value is longer than size limit
// - adding value exceeds cumulative cache capacity
func(st *State) Add(key string, value string, sizeLimit uint16) error {
	if sizeLimit > 0 {
		l := uint16(len(value))
		if l > sizeLimit {
			return fmt.Errorf("value length %v exceeds value size limit %v", l, sizeLimit)
		}
	}
	checkFrame := st.frameOf(key)
	if checkFrame > -1 {
		if checkFrame == len(st.execPath) - 1 {
			log.Printf("Ignoring load request on frame that has symbol already loaded")
			return nil
		}
		return fmt.Errorf("key %v already defined in frame %v", key, checkFrame)
	}
	sz := st.checkCapacity(value)
	if sz == 0 {
		return fmt.Errorf("Cache capacity exceeded %v of %v", st.CacheUseSize + sz, st.CacheSize)
	}
	log.Printf("add key %s value size %v", key, sz)
	st.Cache[len(st.Cache)-1][key] = value
	st.CacheUseSize += sz
	st.sizes[key] = sizeLimit
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
func(st *State) Update(key string, value string) error {
	sizeLimit := st.sizes[key]
	if st.sizes[key] > 0 {
		l := uint16(len(value))
		if l > sizeLimit {
			return fmt.Errorf("update value length %v exceeds value size limit %v", l, sizeLimit)
		}
	}
	checkFrame := st.frameOf(key)
	if checkFrame == -1 {
		return fmt.Errorf("key %v not defined", key)
	}
	r := st.Cache[checkFrame][key]
	l := uint32(len(r))
	st.Cache[checkFrame][key] = ""
	if st.CacheMap[key] != "" {
		st.CacheMap[key] = value
	}
	st.CacheUseSize -= l
	sz := st.checkCapacity(value)
	if sz == 0 {
		baseUseSize := st.CacheUseSize
		st.Cache[checkFrame][key] = r
		st.CacheUseSize += l
		return fmt.Errorf("Cache capacity exceeded %v of %v", baseUseSize + sz, st.CacheSize)
	}
	return nil
}

// Map marks the given key for retrieval.
//
// After this, Val() will return the value for the key, and Size() will include the value size and limitations in its calculations.
//
// Only one symbol with no size limitation may be mapped at the current level.
func(st *State) Map(key string) error {
	m, err := st.Get()
	if err != nil {
		return err
	}
	l := st.sizes[key]
	if l == 0 {
		if st.sink != nil {
			return fmt.Errorf("sink already set to symbol '%v'", *st.sink)
		}
		st.sink = &key
	}
	st.CacheMap[key] = m[key]
	return nil
}

// Depth returns the current call stack depth.
func(st *State) Depth() uint8 {
	return uint8(len(st.Cache))
}

// Get returns the full key-value mapping for all mapped keys at the current cache level.
func(st *State) Get() (map[string]string, error) {
	if len(st.Cache) == 0 {
		return nil, fmt.Errorf("get at top frame")
	}
	return st.Cache[len(st.Cache)-1], nil
}

// Val returns value for key
//
// Fails if key is not mapped.
func(st *State) Val(key string) (string, error) {
	r := st.CacheMap[key]
	if len(r) == 0 {
		return "", fmt.Errorf("key %v not mapped", key)
	}
	return r, nil
}

// Reset flushes all state contents below the top level, and returns to the top level.
func(st *State) Reset() {
	if len(st.Cache) == 0 {
		return
	}
	st.Cache = st.Cache[:1]
	st.CacheUseSize = 0
	return
}

// Check returns true if a key already exists in the cache.
func(st *State) Check(key string) bool {
	return st.frameOf(key) == -1
}

// Size returns size used by values, and remaining size available
func(st *State) Size() (uint32, uint32) {
	var l int
	var c uint16
	for k, v := range st.CacheMap {
		l += len(v)
		c += st.sizes[k]
	}
	r := uint32(l)
	return r, uint32(c)-r
}

// Appendcode adds the given bytecode to the end of the existing code.
func(st *State) AppendCode(b []byte) error {
	st.code = append(st.code, b...)
	log.Printf("code changed to 0x%x", b)
	return nil
}

// SetCode replaces the current bytecode with the given bytecode.
func(st *State) SetCode(b []byte) {
	log.Printf("code set to 0x%x", b)
	st.code = b
}

// Get the remaning cached bytecode
func(st *State) GetCode() ([]byte, error) {
	b := st.code
	st.code = []byte{}
	return b, nil
}

// GetInput gets the most recent client input.
func(st *State) GetInput() ([]byte, error) {
	if st.input == nil {
		return nil, fmt.Errorf("no input has been set")
	}
	return st.input, nil
}

// SetInput is used to record the latest client input.
func(st *State) SetInput(input []byte) error {
	l := len(input)
	if l > 255 {
		return fmt.Errorf("input size %v too large (limit %v)", l, 255)
	}
	st.input = input
	return nil
}

// return 0-indexed frame number where key is defined. -1 if not defined
func(st *State) frameOf(key string) int {
	for i, m := range st.Cache {
		for k, _ := range m {
			if k == key {
				return i
			}
		}
	}
	return -1
}

// bytes that will be added to cache use size for string
// returns 0 if capacity would be exceeded
func(st *State) checkCapacity(v string) uint32 {
	sz := uint32(len(v))
	if st.CacheSize == 0 {
		return sz
	}
	if st.CacheUseSize + sz > st.CacheSize {
		return 0	
	}
	return sz
}

// flush relveant properties for level change
func(st *State) resetCurrent() {
	st.sink = nil
	st.CacheMap = make(map[string]string)
}
