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
}

func NewState(bitSize uint64) State {
	if bitSize == 0 {
		panic("bitsize cannot be 0")
	}
	n := bitSize % 8
	if n > 0 {
		bitSize += (8 - n)
	}

	return State{
		Flags: make([]byte, bitSize / 8),
		CacheSize: 0,
		CacheUseSize: 0,
	}
}

func(st State) WithCacheSize(cacheSize uint32) State {
	st.CacheSize = cacheSize
	return st
}

func(st *State) Enter(input string) {
	m := make(map[string]string)
	st.Cache = append(st.Cache, m)
	st.CacheMap = make(map[string]string)
}

func(st *State) Add(key string, value string) error {
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

func(st *State) Exit() error {
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
	return nil
}

func(st *State) Reset() error {
	st.Cache = st.Cache[:1]
	st.CacheUseSize = 0
	return nil
}

func(st *State) Check(key string) bool {
	return st.frameOf(key) == -1
}

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
