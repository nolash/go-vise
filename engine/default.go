package engine

import (
	"context"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

// NewDefaultEngine is a convenience function to instantiate a filesystem-backed engine with no output constraints.
func NewDefaultEngine(dir string, persisted bool, session *string) (EngineIsh, error) {
	var err error
	st := state.NewState(0)
	rs := resource.NewFsResource(dir)
	ca := cache.NewCache()
	cfg := Config{
		Root: "root",
	}
	if session != nil {
		cfg.SessionId = *session
	} else if !persisted {
		return nil, fmt.Errorf("session must be set if persist is used")	
	}
	ctx := context.TODO()
	var en EngineIsh
	if persisted {
		dp := path.Join(dir, ".state")
		err = os.MkdirAll(dp, 0700)
		if err != nil {
			return nil, err
		}
		pr := persist.NewFsPersister(dp)
		en, err = NewPersistedEngine(cfg, pr, &rs, ctx)
		if err != nil {
			Logg.Infof("persisted engine create error. trying again with persisting empty state first...")
			pr = pr.WithContent(&st, ca)
			err = pr.Save(cfg.SessionId, nil)
			if err != nil {
				return nil, err
			}
			en, err = NewPersistedEngine(cfg, pr, &rs, ctx)
		}
	} else {
		enb := NewEngine(cfg, &st, &rs, ca, ctx)
		en = &enb
	}
	return en, err
}

// NewSizedEngine is a convenience function to instantiate a filesystem-backed engine with a specified output constraint.
func NewSizedEngine(dir string, size uint32, persisted bool, session *string) (EngineIsh, error) {
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
	} else if !persisted {
		return nil, fmt.Errorf("session must be set if persist is used")
	}
	ctx := context.TODO()
	var en EngineIsh
	if persisted {
		dp := path.Join(dir, ".state")
		err = os.MkdirAll(dp, 0700)
		if err != nil {
			return nil, err
		}
		pr := persist.NewFsPersister(dp)
		en, err = NewPersistedEngine(cfg, pr, &rs, ctx)
		if err != nil {
			Logg.Infof("persisted engine create error. trying again with persisting empty state first...")
			pr = pr.WithContent(&st, ca)
			err = pr.Save(cfg.SessionId, nil)
			if err != nil {
				return nil, err
			}
			en, err = NewPersistedEngine(cfg, pr, &rs, ctx)
		}
	} else {
		enb := NewEngine(cfg, &st, &rs, ca, ctx)
		en = &enb
	}
	return en, err
}
