package state

import (
	"fmt"
	"log"
)

type State struct {
	Flags []byte
	CacheSize uint32
	CacheUseSize uint32
	Cache []map[string]string
	CacheMap map[string]string
	ExecPath []string
	Arg *string
	Idx uint16
}

func NewState(bitSize uint64) State {
	if bitSize == 0 {
		panic("bitsize cannot be 0")
	}
	n := bitSize % 8
	if n > 0 {
		bitSize += (8 - n)
	}

	st := State{
		Flags: make([]byte, bitSize / 8),
		CacheSize: 0,
		CacheUseSize: 0,
	}
	st.Down("")
	return st
}

func(st State) Where() string {
	if len(st.ExecPath) == 0 {
		return ""
	}
	l := len(st.ExecPath)
	return st.ExecPath[l-1]
}

func(st State) WithCacheSize(cacheSize uint32) State {
	st.CacheSize = cacheSize
	return st
}

func(st *State) PutArg(input string) error {
	st.Arg = &input
	return nil
}

func(st *State) PopArg() (string, error) {
	if st.Arg == nil {
		return "", fmt.Errorf("arg is not set")
	}
	return *st.Arg, nil
}

func(st *State) Down(input string) {
	m := make(map[string]string)
	st.Cache = append(st.Cache, m)
	st.CacheMap = make(map[string]string)
	st.ExecPath = append(st.ExecPath, input)
}

func(st *State) Add(key string, value string, sizeHint uint32) error {
	checkFrame := st.frameOf(key)
	if checkFrame > -1 {
		return fmt.Errorf("key %v already defined in frame %v", key, checkFrame)
	}
	sz := st.checkCapacity(value)
	if sz == 0 {
		return fmt.Errorf("Cache capacity exceeded %v of %v", st.CacheUseSize + sz, st.CacheSize)
	}
	log.Printf("add key %s value size %v", key, sz)
	st.Cache[len(st.Cache)-1][key] = value
	st.CacheUseSize += sz
	_ = sizeHint
	return nil
}

func(st *State) Update(key string, value string) error {
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

func(st *State) Map(k string) error {
	m, err := st.Get()
	if err != nil {
		return err
	}
	st.CacheMap[k] = m[k]
	return nil
}

func(st *State) Depth() uint8 {
	return uint8(len(st.Cache))
}

func(st *State) Get() (map[string]string, error) {
	if len(st.Cache) == 0 {
		return nil, fmt.Errorf("get at top frame")
	}
	return st.Cache[len(st.Cache)-1], nil
}

func(st *State) Val(key string) (string, error) {
	r := st.CacheMap[key]
	if len(r) == 0 {
		return "", fmt.Errorf("key %v not mapped", key)
	}
	return r, nil
}

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
	st.ExecPath = st.ExecPath[:l]
	return nil
}

func(st *State) Reset() {
	if len(st.Cache) == 0 {
		return
	}
	st.Cache = st.Cache[:1]
	st.CacheUseSize = 0
	return
}

func(st *State) Check(key string) bool {
	return st.frameOf(key) == -1
}

// return 0-indexed frame number where key is defined. -1 if not defined
func(st *State) frameOf(key string) int {
	log.Printf("--- %s", key)
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
