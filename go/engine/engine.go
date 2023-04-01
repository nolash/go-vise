package engine

import (
	"context"
	"fmt"
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
	st *state.State
	rs resource.Resource
}

func NewEngine(st *state.State, rs resource.Resource) Engine {
	engine := Engine{st, rs}
	return engine
}

func(en *Engine) Init(ctx context.Context) error {
	b := vm.NewLine([]byte{}, vm.MOVE, []string{"root"}, nil, nil)
	var err error
	_, err = vm.Run(b, en.st, en.rs, ctx)
	if err != nil {
		return err
	}
	location := en.st.Where()
	code, err := en.rs.GetCode(location)
	if err != nil {
		return err
	}
	return en.st.AppendCode(code)
}

func (en *Engine) Exec(input []byte, ctx context.Context) error {
	l := uint8(len(input))
	if l > 255 {
		return fmt.Errorf("input too long (%v)", l)
	}
	input = append([]byte{l}, input...)
	code, err := en.st.GetCode()
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return fmt.Errorf("no code to execute")
	}
	code, err = vm.Apply(input, code, en.st, en.rs, ctx)
	en.st.SetCode(code)
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
