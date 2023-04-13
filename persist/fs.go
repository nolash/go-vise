package persist

import (
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"github.com/fxamacker/cbor/v2"

	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/state"
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

// GetState implements the Persister interface.
func(p *FsPersister) GetMemory() cache.Memory {
	return p.Memory
}

// GetState implements the Persister interface.
func(p *FsPersister) Serialize() ([]byte, error) {
	return cbor.Marshal(p)
}

// GetState implements the Persister interface.
func(p *FsPersister) Deserialize(b []byte) error {
	err := cbor.Unmarshal(b, p)
	return err
}

// GetState implements the Persister interface.
func(p *FsPersister) Save(key string) error {
	b, err := p.Serialize()
	if err != nil {
		return err
	}
	fp := path.Join(p.dir, key)
	log.Printf("saved key %v state %x", key, p.State.Code)
	return ioutil.WriteFile(fp, b, 0600)
}

// GetState implements the Persister interface.
func(p *FsPersister) Load(key string) error {
	fp := path.Join(p.dir, key)
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	err = p.Deserialize(b)
	log.Printf("loaded key %v state %x", key, p.State.Code)
	return err
}
