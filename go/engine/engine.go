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
// It loads and executes code for the start node.
func(en *Engine) Init(sym string, ctx context.Context) error {
	err := en.st.SetInput([]byte{})
	if err != nil {
		return err
	}
	b := vm.NewLine(nil, vm.MOVE, []string{sym}, nil, nil)
	b, err = vm.Run(b, en.st, en.rs, ctx)
	if err != nil {
		return err
	}
	en.st.SetCode(b)
	return nil
}

// Exec processes user input against the current state of the virtual machine environment.
//
// If successfully executed, output of the last execution is available using the WriteResult call.
// 
// A bool return valus of false indicates that execution should be terminated. Calling Exec again has undefined effects.
//
// Fails if:
// - input is formally invalid (too long etc)
// - no current bytecode is available
// - input processing against bytcode failed
func (en *Engine) Exec(input []byte, ctx context.Context) (bool, error) {
	err := vm.ValidInput(input)
	if err != nil {
		return true, err
	}
	err = en.st.SetInput(input)
	if err != nil {
		return false, err
	}

	log.Printf("new execution with input '%s' (0x%x)", input, input)
	code, err := en.st.GetCode()
	if err != nil {
		return false, err
	}
	if len(code) == 0 {
		return false, fmt.Errorf("no code to execute")
	}
	code, err = vm.Run(code, en.st, en.rs, ctx)
	if err != nil {
		return false, err
	}

	v, err := en.st.MatchFlag(state.FLAG_TERMINATE, false)
	if err != nil {
		return false, err
	}
	if v {
		if len(code) > 0 {
			log.Printf("terminated with code remaining: %x", code)
		}
		return false, nil
	}

	en.st.SetCode(code)
	if len(code) == 0 {
		log.Printf("runner finished with no remaining code")
		return false, nil
	}

	return true, nil
}

// WriteResult writes the output of the last vm execution to the given writer.
//
// Fails if
// - required data inputs to the template are not available.
// - the template for the given node point is note available for retrieval using the resource.Resource implementer.
// - the supplied writer fails to process the writes.
func(en *Engine) WriteResult(w io.Writer) error {
	location, idx := en.st.Where()
	v, err := en.st.Get()
	if err != nil {
		return err
	}
	r, err := en.rs.RenderTemplate(location, v, idx, nil)
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
