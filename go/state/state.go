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
}

func(st *State) Add(k string, v string) error {
	sz := st.checkCapacity(v)
	if sz == 0 {
		return fmt.Errorf("Cache capacity exceeded %v of %v", st.CacheUseSize + sz, st.CacheSize)
	}
	log.Printf("add key %s value size %v", k, sz)
	st.Cache[len(st.Cache)-1][k] = v
	st.CacheUseSize += sz
	return nil
}

func(st *State) Get() (map[string]string, error) {
	return st.Cache[len(st.Cache)-1], nil
}

func (st *State) Exit() error {
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
