package engine

import (
	"context"
	"fmt"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/db"
)

// NewDefaultEngine is a convenience function to instantiate a filesystem-backed engine with no output constraints.
func NewDefaultEngine(dir string, persistDb db.Db, session *string) (EngineIsh, error) {
	var err error
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	cfg := Config{
		Root: "root",
	}
	if session != nil {
		cfg.SessionId = *session
	} else if persistDb != nil {
		return nil, fmt.Errorf("session must be set if persist is used")	
	}
	ctx := context.TODO()
	var en EngineIsh
	if persistDb != nil {
		pr := persist.NewPersister(persistDb)
		en, err = NewPersistedEngine(ctx, cfg, pr, rs)
		if err != nil {
			logg.Infof("persisted engine create error. trying again with persisting empty state first...")
			pr = pr.WithContent(&st, ca)
			err = pr.Save(cfg.SessionId)
			if err != nil {
				return nil, err
			}
			en, err = NewPersistedEngine(ctx, cfg, pr, rs)
		}
	} else {
		enb := NewEngine(ctx, cfg, &st, rs, ca)
		en = &enb
	}
	return en, err
}

// NewSizedEngine is a convenience function to instantiate a filesystem-backed engine with a specified output constraint.
func NewSizedEngine(dir string, size uint32, persistDb db.Db, session *string) (EngineIsh, error) {
	var err error
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	cfg := Config{
		OutputSize: size,
		Root: "root",
	}
	if session != nil {
		cfg.SessionId = *session
	} else if persistDb != nil {
		return nil, fmt.Errorf("session must be set if persist is used")
	}
	ctx := context.TODO()
	var en EngineIsh
	if persistDb != nil {
		pr := persist.NewPersister(persistDb)
		en, err = NewPersistedEngine(ctx, cfg, pr, rs)
		if err != nil {
			logg.Infof("persisted engine create error. trying again with persisting empty state first...")
			pr = pr.WithContent(&st, ca)
			err = pr.Save(cfg.SessionId)
			if err != nil {
				return nil, err
			}
			en, err = NewPersistedEngine(ctx, cfg, pr, rs)
		}
	} else {
		enb := NewEngine(ctx, cfg, &st, rs, ca)
		en = &enb
	}
	return en, err
}
