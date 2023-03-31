package state

import (
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
	st.Cache[len(st.Cache)-1][k] = v
	return nil
}

func(st *State) Get() (map[string]string, error) {
	return st.Cache[len(st.Cache)-1], nil
}
