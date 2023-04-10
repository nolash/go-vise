package engine

import (
	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
)

func NewDefaultEngine(dir string) Engine {
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	return NewEngine(Config{}, &st, &rs, ca)
}

func NewSizedEngine(dir string, size uint32) Engine {
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	cfg := Config{
		OutputSize: size,
	}
	return NewEngine(cfg, &st, &rs, ca)
}
