package persist

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"github.com/fxamacker/cbor/v2"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
)

// FsPersister is an implementation of Persister that saves state to the file system.
type FsPersister struct {
	State *state.State
	Memory *cache.Cache
	dir string
}

// NewFsPersister creates a new FsPersister.
//
// The filesystem store will be at the given directory. The directory must exist.
func NewFsPersister(dir string) *FsPersister {
	fp, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	return &FsPersister{
		dir: fp,
	}
}

// WithContent sets a current State and Cache object.
//
// This method is normally called before Serialize / Save.
func(p *FsPersister) WithContent(st *state.State, ca *cache.Cache) *FsPersister {
	p.State = st
	p.Memory = ca
	return p
}

// GetState implements the Persister interface.
func(p *FsPersister) GetState() *state.State {
	return p.State
}

// GetMemory implements the Persister interface.
func(p *FsPersister) GetMemory() cache.Memory {
	return p.Memory
}

// Serialize implements the Persister interface.
func(p *FsPersister) Serialize() ([]byte, error) {
	return cbor.Marshal(p)
}

// Deserialize implements the Persister interface.
func(p *FsPersister) Deserialize(b []byte) error {
	err := cbor.Unmarshal(b, p)
	return err
}

// Save implements the Persister interface.
func(p *FsPersister) Save(key string) error {
	b, err := p.Serialize()
	if err != nil {
		return err
	}
	fp := path.Join(p.dir, key)
	Logg.Debugf("saved state and cache", "key", key, "bytecode", p.State.Code, "flags", p.State.Flags)
	return ioutil.WriteFile(fp, b, 0600)
}

// Load implements the Persister interface.
func(p *FsPersister) Load(key string) error {
	fp := path.Join(p.dir, key)
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	err = p.Deserialize(b)
	Logg.Debugf("loaded state and cache", "key", key, "bytecode", p.State.Code)
	return err
}
