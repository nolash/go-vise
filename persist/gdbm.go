package persist

import (
	"github.com/fxamacker/cbor/v2"
	gdbm "github.com/graygnuorg/go-gdbm"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/db"
)

// gdbmPersister is an implementation of Persister that saves state to the file system.
type gdbmPersister struct {
	State *state.State
	Memory *cache.Cache
	db *gdbm.Database
}

func NewGdbmPersiser(fp string) *gdbmPersister {
	gdb, err := gdbm.Open(fp, gdbm.ModeReader)
	if err != nil {
		panic(err)
	}
	return NewGdbmPersisterFromDatabase(gdb)
}

func NewGdbmPersisterFromDatabase(gdb *gdbm.Database) *gdbmPersister {
	return &gdbmPersister{
		db: gdb,
	}
}

// WithContent sets a current State and Cache object.
//
// This method is normally called before Serialize / Save.
func(p *gdbmPersister) WithContent(st *state.State, ca *cache.Cache) *gdbmPersister {
	p.State = st
	p.Memory = ca
	return p
}

// TODO: DRY
// GetState implements the Persister interface.
func(p *gdbmPersister) GetState() *state.State {
	return p.State
}

// GetMemory implements the Persister interface.
func(p *gdbmPersister) GetMemory() cache.Memory {
	return p.Memory
}

// Serialize implements the Persister interface.
func(p *gdbmPersister) Serialize() ([]byte, error) {
	return cbor.Marshal(p)
}

// Deserialize implements the Persister interface.
func(p *gdbmPersister) Deserialize(b []byte) error {
	err := cbor.Unmarshal(b, p)
	return err
}

// Save implements the Persister interface.
func(p *gdbmPersister) Save(key string) error {
	b, err := p.Serialize()
	if err != nil {
		return err
	}
	k := db.ToDbKey(db.DATATYPE_STATE, []byte(key), nil)
	err = p.db.Store(k, b, true)
	if err != nil {
		return err
	}
	Logg.Debugf("saved state and cache", "key", key, "bytecode", p.State.Code, "flags", p.State.Flags)
	return nil
}

// Load implements the Persister interface.
func(p *gdbmPersister) Load(key string) error {
	k := db.ToDbKey(db.DATATYPE_STATE, []byte(key), nil)
	b, err := p.db.Fetch(k)
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
