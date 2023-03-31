package state

import (
	"io"
)

type State struct {
	Flags []byte
	OutputSize uint16
	CacheSize uint32
	CacheUseSize uint32
	Cache io.ReadWriteSeeker
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
		Cache: nil,
	}
}

func(st State) WithCacheSize(cacheSize uint32) State {
	st.CacheSize = cacheSize
	return st
}
