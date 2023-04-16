package engine

import (
	"context"
	"fmt"
	"io"
	"log"

	"git.defalsify.org/vise/cache"
	"git.defalsify.org/vise/render"
	"git.defalsify.org/vise/resource"
	"git.defalsify.org/vise/state"
	"git.defalsify.org/vise/vm"
)

// Config globally defines behavior of all components driven by the engine.
type Config struct {
	OutputSize uint32 // Maximum size of output from a single rendered page
	SessionId string
	Root string
	FlagCount uint32
	CacheSize uint32
}

// Engine is an execution engine that handles top-level errors when running client inputs against code in the bytecode buffer.
type Engine struct {
	st *state.State
	rs resource.Resource
	ca cache.Memory
	vm *vm.Vm
	root string
	initd bool
}

// NewEngine creates a new Engine
func NewEngine(cfg Config, st *state.State, rs resource.Resource, ca cache.Memory, ctx context.Context) (Engine, error) {
	var szr *render.Sizer
	if cfg.OutputSize > 0 {
		szr = render.NewSizer(cfg.OutputSize)
	}
	ctx = context.WithValue(ctx, "sessionId", cfg.SessionId)
	engine := Engine{
		st: st,
		rs: rs,
		ca: ca,
		vm: vm.NewVm(st, rs, ca, szr),
	}
	engine.root = cfg.Root	
	return engine, nil
}

// Init must be explicitly called before using the Engine instance.
//
// It loads and executes code for the start node.
func(en *Engine) Init(sym string, ctx context.Context) error {
	if en.initd {
		log.Printf("already initialized")
		return nil
	}
	if sym == "" {
		return fmt.Errorf("start sym empty")
	}
	inSave, _ := en.st.GetInput()
	err := en.st.SetInput([]byte{})
	if err != nil {
		return err
	}
	b := vm.NewLine(nil, vm.MOVE, []string{sym}, nil, nil)
	log.Printf("start new init VM run with code %x", b)
	b, err = en.vm.Run(b, ctx)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return fmt.Errorf("no code left after init, that's just useless and sad")
	}
	log.Printf("ended init VM run with code %x", b)
	en.st.SetCode(b)
	err = en.st.SetInput(inSave)
	if err != nil {
		return err
	}
	en.initd = true
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
	var err error
	if en.st.Moves == 0 {
		err = en.Init(en.root, ctx)
		if err != nil {
			return false, err
		}
		if len(input) == 0 {
			return true, nil
		}
	}
	err = vm.ValidInput(input)
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

	log.Printf("start new VM run with code %x", code)
	code, err = en.vm.Run(code, ctx)
	if err != nil {
		return false, err
	}
	log.Printf("ended VM run with code %x", code)

	v, err := en.st.MatchFlag(state.FLAG_TERMINATE, false)
	if err != nil {
		return false, err
	}
	if v {
		if len(code) > 0 {
			log.Printf("terminated with code remaining: %x", code)
		}
		return false, err
	}

	en.st.SetCode(code)
	if len(code) == 0 {
		log.Printf("runner finished with no remaining code")
		err = en.reset(ctx)
		return false, err
	}

	return true, nil
}

// WriteResult writes the output of the last vm execution to the given writer.
//
// Fails if
// - required data inputs to the template are not available.
// - the template for the given node point is note available for retrieval using the resource.Resource implementer.
// - the supplied writer fails to process the writes.
func(en *Engine) WriteResult(w io.Writer, ctx context.Context) (int, error) {
	r, err := en.vm.Render(ctx)
	if err != nil {
		return 0, err
	}
	return io.WriteString(w, r)
}

func(en *Engine) reset(ctx context.Context) error {
	var err error
	var isTop bool
	for !isTop {
		isTop, err = en.st.Top()
		if err != nil {
			return err
		}
		_, err = en.st.Up()
		if err != nil {
			return err
		}
		en.ca.Pop()
	}
	en.st.Restart()
	en.initd = false
	return en.Init(en.root, ctx)
}
