package persist

import (
	"context"

	"github.com/fxamacker/cbor/v2"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/cache"
)

type Persister struct {
	State *state.State
	Memory *cache.Cache
	ctx context.Context
	db db.Db
}

func NewPersister(db db.Db) *Persister {
	return &Persister{
		db: db,
		ctx: context.Background(),
	}
}

func(p *Persister) WithContext(ctx context.Context) *Persister {
	p.ctx = ctx
	return p
}

func(p *Persister) WithSession(sessionId string) *Persister {
	p.db.SetSession(sessionId)
	return p
}

// WithContent sets a current State and Cache object.
//
// This method is normally called before Serialize / Save.
func(p *Persister) WithContent(st *state.State, ca *cache.Cache) *Persister {
	p.State = st
	p.Memory = ca
	return p
}

// GetState implements the Persister interface.
func(p *Persister) GetState() *state.State {
	return p.State
}

// GetMemory implements the Persister interface.
func(p *Persister) GetMemory() cache.Memory {
	return p.Memory
}

// Serialize implements the Persister interface.
func(p *Persister) Serialize() ([]byte, error) {
	return cbor.Marshal(p)
}

// Deserialize implements the Persister interface.
func(p *Persister) Deserialize(b []byte) error {
	err := cbor.Unmarshal(b, p)
	return err
}

func(p *Persister) Save(key string) error {
	b, err := p.Serialize()
	if err != nil {
		return err
	}
	p.db.SetPrefix(db.DATATYPE_STATE)
	return p.db.Put(p.ctx, []byte(key), b)
}

func(p *Persister) Load(key string) error {
	p.db.SetPrefix(db.DATATYPE_STATE)
	b, err := p.db.Get(p.ctx, []byte(key))
	if err != nil {
		return err
	}
	err = p.Deserialize(b)
	if err != nil {
		return err
	}
	Logg.Debugf("loaded state and cache", "key", key, "bytecode", p.State.Code)
	return nil
}
