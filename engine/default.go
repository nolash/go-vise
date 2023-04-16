package engine

import (
	"context"

	"git.defalsify.org/vise/cache"
	"git.defalsify.org/vise/resource"
	"git.defalsify.org/vise/state"
)

// NewDefaultEngine is a convenience function to instantiate a filesystem-backed engine with no output constraints.
func NewDefaultEngine(dir string) (Engine, error) {
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
func NewSizedEngine(dir string, size uint32) (Engine, error) {
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
