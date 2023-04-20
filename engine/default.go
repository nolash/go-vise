package engine

import (
	"context"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

// NewDefaultEngine is a convenience function to instantiate a filesystem-backed engine with no output constraints.
func NewDefaultEngine(dir string) Engine {
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	cfg := Config{
		Root: "root",
	}
	ctx := context.TODO()
	return NewEngine(cfg, &st, &rs, ca, ctx)
}

// NewSizedEngine is a convenience function to instantiate a filesystem-backed engine with a specified output constraint.
func NewSizedEngine(dir string, size uint32) Engine {
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	cfg := Config{
		OutputSize: size,
		Root: "root",
	}
	ctx := context.TODO()
	return NewEngine(cfg, &st, &rs, ca, ctx)
}
