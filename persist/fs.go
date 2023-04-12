package persist

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"github.com/fxamacker/cbor/v2"

	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/state"
)

type FsPersister struct {
	State *state.State
	Memory *cache.Cache
	dir string
}

func NewFsPersister(dir string) *FsPersister {
	fp, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	return &FsPersister{
		dir: fp,
	}
}

func(p *FsPersister) WithContent(st *state.State, ca *cache.Cache) *FsPersister {
	p.State = st
	p.Memory = ca
	return p
}

func(p *FsPersister) Serialize() ([]byte, error) {
	return cbor.Marshal(p)
}

func(p *FsPersister) Deserialize(b []byte) error {
	err := cbor.Unmarshal(b, p)
	return err
}

func(p *FsPersister) Save(key string) error {
	b, err := p.Serialize()
	if err != nil {
		return err
	}
	fp := path.Join(p.dir, key)
	return ioutil.WriteFile(fp, b, 0600)
}

func(p *FsPersister) Load(key string) error {
	fp := path.Join(p.dir, key)
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	err = p.Deserialize(b)
	return err
}
