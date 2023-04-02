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

// Engine is an execution engine that handles top-level errors when running user inputs against currently exposed bytecode.
type Engine struct {
	st *state.State
	rs resource.Resource
}

// NewEngine creates a new Engine
func NewEngine(st *state.State, rs resource.Resource) Engine {
	engine := Engine{st, rs}
	return engine
}

// Init must be explicitly called before using the Engine instance.
//
// It makes sure bootstrapping code has been executed, and that the exposed bytecode is ready for user input.
func(en *Engine) Init(sym string, ctx context.Context) error {
	b := vm.NewLine(nil, vm.MOVE, []string{sym}, nil, nil)
	var err error
	b, err = vm.Run(b, en.st, en.rs, ctx)
	if err != nil {
		return err
	}
//	location := en.st.Where()
//	code, err := en.rs.GetCode(location)
//	if err != nil {
//		return err
//	}
//	if len(code) == 0 {
//		return fmt.Errorf("no code found at resource %s", en.rs)
//	}
//
//	code, err = vm.Run(code, en.st, en.rs, ctx)
//
	en.st.SetCode(b)
	return nil
}

// Exec processes user input against the current state of the virtual machine environment.
//
// If successfully executed:
// - output of the last execution is available using the WriteResult(...) call
// - Exec(...) may be called again with new input
//
// This implementation is in alpha state. That means that any error emitted may have left the system in an undefined state.
//
// TODO: Disambiguate errors as critical and resumable errors.
//
// Fails if:
// - input is objectively invalid (too long etc)
// - no current bytecode is available
// - input processing against bytcode failed
func (en *Engine) Exec(input []byte, ctx context.Context) error {
	err := en.st.SetInput(input)
	if err != nil {
		return err
	}
	log.Printf("new execution with input 0x%x (%v)", input, len(input))
	code, err := en.st.GetCode()
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return fmt.Errorf("no code to execute")
	}
	code, err = vm.Run(code, en.st, en.rs, ctx)
	en.st.SetCode(code)
	return err
}

// WriteResult writes the output of the last vm execution to the given writer.
//
// Fails if
// - required data inputs to the template are not available.
// - the template for the given node point is note available for retrieval using the resource.Resource implementer.
// - the supplied writer fails to process the writes.
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
	m, err := en.rs.RenderMenu()
	if err != nil {
		return err
	}
	if len(m) > 0 {
		r += "\n" + m
	}
	c, err := io.WriteString(w, r)
	log.Printf("%v bytes written as result for %v", c, location)
	return err
}
