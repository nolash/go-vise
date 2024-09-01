package engine

import (
	"context"
	"fmt"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

// NewDefaultEngine is a convenience function to instantiate a filesystem-backed engine with no output constraints.
func NewDefaultEngine(dir string, persistDb db.Db, session *string) (EngineIsh, error) {
	st := state.NewState(0)
	ctx := context.Background()
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, dir)
	if err != nil {
		return nil, err
	}
	rs := resource.NewDbResource(store)
	rs.With(db.DATATYPE_STATICLOAD)
	ca := cache.NewCache()
	cfg := Config{
		Root: "root",
	}
	if session != nil {
		cfg.SessionId = *session
	} else if persistDb != nil {
		return nil, fmt.Errorf("session must be set if persist is used")	
	}
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
	st := state.NewState(0)
	ca := cache.NewCache()
	ctx := context.Background()
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, dir)
	if err != nil {
		return nil, err
	}
	rs := resource.NewDbResource(store)
	rs.With(db.DATATYPE_STATICLOAD)
	cfg := Config{
		OutputSize: size,
		Root: "root",
	}
	if session != nil {
		cfg.SessionId = *session
	} else if persistDb != nil {
		return nil, fmt.Errorf("session must be set if persist is used")
	}
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
