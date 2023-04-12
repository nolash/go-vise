package engine

import (
	"context"
	"io"
	"log"

	"git.defalsify.org/festive/persist"
	"git.defalsify.org/festive/resource"
)

func RunPersisted(cfg Config, rs resource.Resource, pr persist.Persister, input []byte, w io.Writer, ctx context.Context) error {
	err := pr.Load(cfg.SessionId)
	if err != nil {
		return err
	}
	st := pr.GetState()
	log.Printf("st %v", st)
	en := NewEngine(cfg, pr.GetState(), rs, pr.GetMemory(), ctx)

	if len(input) > 0 {
		_, err = en.Exec(input, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
