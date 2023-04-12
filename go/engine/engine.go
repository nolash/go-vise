package engine

import (
	"context"
	"fmt"
	"io"
	"log"

	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/render"
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/vm"
)

type Config struct {
	OutputSize uint32
//	FlagCount uint32
//	CacheSize uint32
}

// Engine is an execution engine that handles top-level errors when running user inputs against currently exposed bytecode.
type Engine struct {
	st *state.State
	rs resource.Resource
	ca cache.Memory
	vm *vm.Vm
}

// NewEngine creates a new Engine
func NewEngine(cfg Config, st *state.State, rs resource.Resource, ca cache.Memory) Engine {
	var szr *render.Sizer
	if cfg.OutputSize > 0 {
		szr = render.NewSizer(cfg.OutputSize)
	}
	engine := Engine{
		st: st,
		rs: rs,
		ca: ca,
		vm: vm.NewVm(st, rs, ca, szr),
	}
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
	b, err = en.vm.Run(b, ctx)
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
	code, err = en.vm.Run(code, ctx)
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
func(en *Engine) WriteResult(w io.Writer, ctx context.Context) error {
	r, err := en.vm.Render(ctx)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, r)
	return err
}
