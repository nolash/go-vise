package engine

import (
	"context"
	"io"

	"git.defalsify.org/festive/persist"
	"git.defalsify.org/festive/resource"
)

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
func RunPersisted(cfg Config, rs resource.Resource, pr persist.Persister, input []byte, w io.Writer, ctx context.Context) error {
	err := pr.Load(cfg.SessionId)
	if err != nil {
		return err
	}

	en := NewEngine(cfg, pr.GetState(), rs, pr.GetMemory(), ctx)

	c, err := en.WriteResult(w, ctx)
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

	_, err = en.Exec(input, ctx)
	if err != nil {
		return err
	}
	_, err = en.WriteResult(w, ctx)
	if err != nil {
		return err
	}
	return pr.Save(cfg.SessionId)
}
