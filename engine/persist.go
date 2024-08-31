package engine

import (
	"context"
	"io"

	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
)

// PersistedEngine adds persisted state to the Engine object. It provides a persisted state option for synchronous/interactive clients.
type PersistedEngine struct {
	*Engine
	pr *persist.Persister
}


// NewPersistedEngine creates a new PersistedEngine
func NewPersistedEngine(ctx context.Context, cfg Config, pr *persist.Persister, rs resource.Resource) (PersistedEngine, error) {
	err := pr.Load(cfg.SessionId)
	if err != nil {
		return PersistedEngine{}, err
	}
	st := pr.GetState()
	ca := pr.GetMemory()
	
	enb := NewEngine(ctx, cfg, st, rs, ca)
	en := PersistedEngine{
		&enb,
		pr,
	}
	return en, err
}

// Exec executes the parent method Engine.Exec, and afterwards persists the new state.
func(pe PersistedEngine) Exec(ctx context.Context, input []byte) (bool, error) {
	v, err := pe.Engine.Exec(ctx, input)
	if err != nil {
		return v, err
	}
	err = pe.pr.Save(pe.Engine.session)
	return v, err
}

// Finish implements EngineIsh interface
func(pe PersistedEngine) Finish() error {
	logg.Tracef("that's a wrap", "engine", pe)
	return pe.pr.Save(pe.Engine.session)
}

// RunPersisted performs a single vm execution from client input using a persisted state.
//
// State is first loaded from storage. The vm is initialized with the state and executed. The new state is then saved to storage.
//
// The resulting output of the execution will be written to the provided writer.
//
// The state is identified by the SessionId member of the Config. Before first execution, the caller must ensure that an
// initialized state actually is available for the identifier, otherwise the method will fail.
//
// It will also fail if execution by the underlying Engine fails.
func RunPersisted(cfg Config, rs resource.Resource, pr *persist.Persister, input []byte, w io.Writer, ctx context.Context) error {
	err := pr.Load(cfg.SessionId)
	if err != nil {
		return err
	}

	st := pr.GetState()
	ca := pr.GetMemory()
	en := NewEngine(ctx, cfg, st, rs, ca)

	c, err := en.WriteResult(ctx, w)
	if err != nil {
		return err
	}
	err = pr.Save(cfg.SessionId)
	if err != nil {
		return err
	}
	if c > 0 {
		return err
	}

	_, err = en.Exec(ctx, input)
	if err != nil {
		return err
	}
	_, err = en.WriteResult(ctx, w)
	if err != nil {
		return err
	}
	en.Finish()
	return pr.Save(cfg.SessionId)
}
