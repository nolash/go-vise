package state

import (
	"fmt"
)

type State struct {
	Flags []byte
	OutputSize uint16
	CacheSize uint32
	CacheUseSize uint32
	Cache []map[string]string
	ExecPath []string
}

func NewState(bitSize uint64, outputSize uint16) State {
	if bitSize == 0 {
		panic("bitsize cannot be 0")
	}
	n := bitSize % 8
	if n > 0 {
		bitSize += (8 - n)
	}

	return State{
		Flags: make([]byte, bitSize / 8),
		OutputSize: outputSize,
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
	fmt.Printf("len %v %v\n", sz, st.CacheUseSize)
	st.Cache[len(st.Cache)-1][k] = v
	st.CacheUseSize += sz
	return nil
}

func(st *State) Get() (map[string]string, error) {
	return st.Cache[len(st.Cache)-1], nil
}

func (st *State) Exit() {
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
