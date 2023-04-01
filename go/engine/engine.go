package engine

import (
	"context"
	"io"
	"log"

	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/vm"
)
//
//type Config struct {
//	FlagCount uint32
//	CacheSize uint32
//}

type Engine struct {
	st state.State
	rs resource.Resource
}

func NewEngine(st state.State, rs resource.Resource) Engine {
	engine := Engine{st, rs}
	return engine
}

func(en *Engine) Init(ctx context.Context) error {
	b := vm.NewLine([]byte{}, vm.MOVE, []string{"root"}, nil, nil)
	var err error
	en.st, _, err = vm.Run(b, en.st, en.rs, ctx)
	return err
}

func(en *Engine) WriteResult(w io.Writer) error {
	location := en.st.Where()
	v, err := en.st.Get()
	if err != nil {
		return err
	}
	r, err := en.rs.RenderTemplate(location, v)
	if err != nil {
		return err
	}
	c, err := io.WriteString(w, r)
	log.Printf("%v bytes written as result for %v", c, location)
	return err
}
