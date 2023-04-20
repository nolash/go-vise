package engine

import (
	"context"
	"fmt"
	"io"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/render"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/vm"
)

type EngineIsh interface {
	Init(ctx context.Context) (bool, error)
	Exec(input []byte, ctx context.Context) (bool, error)
	WriteResult(w io.Writer, ctx context.Context) (int, error)
	Finish() error
}

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
	session string
	initd bool
}

// NewEngine creates a new Engine
func NewEngine(cfg Config, st *state.State, rs resource.Resource, ca cache.Memory, ctx context.Context) Engine {
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
	engine.session = cfg.SessionId

	return engine
}

// Finish implements EngineIsh interface
func(en *Engine) Finish() error {
	return nil
}

func(en *Engine) restore() {
	location, _ := en.st.Where()
	if len(location) == 0 {
		return
	}
	if en.root != location {
		en.root = "." //location
	}
}

// Init must be explicitly called before using the Engine instance.
//
// It loads and executes code for the start node.
func(en *Engine) Init(ctx context.Context) (bool, error) {
	en.restore()
	if en.initd {
		Logg.DebugCtxf(ctx, "already initialized")
		return true, nil
	}
	sym := en.root
	if sym == "" {
		return false, fmt.Errorf("start sym empty")
	}
	inSave, _ := en.st.GetInput()
	err := en.st.SetInput([]byte{})
	if err != nil {
		return false, err
	}
	b := vm.NewLine(nil, vm.MOVE, []string{sym}, nil, nil)
	Logg.DebugCtxf(ctx, "start new init VM run", "code", b)
	b, err = en.vm.Run(b, ctx)
	if err != nil {
		return false, err
	}
	
	Logg.DebugCtxf(ctx, "end new init VM run", "code", b)
	en.st.SetCode(b)
	err = en.st.SetInput(inSave)
	if err != nil {
		return false, err
	}
	return len(b) > 0, nil
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
		cont, err := en.Init(ctx)
		if err != nil {
			return false, err
		}
		return cont, nil
	}
	err = vm.ValidInput(input)
	if err != nil {
		return true, err
	}
	err = en.st.SetInput(input)
	if err != nil {
		return false, err
	}
	return en.exec(input, ctx)
}

func(en *Engine) exec(input []byte, ctx context.Context) (bool, error) {
	Logg.InfoCtxf(ctx, "new VM execution with input", "input", string(input))
	code, err := en.st.GetCode()
	if err != nil {
		return false, err
	}
	if len(code) == 0 {
		return false, fmt.Errorf("no code to execute")
	}

	Logg.Debugf("start new VM run", "code", code)
	code, err = en.vm.Run(code, ctx)
	if err != nil {
		return false, err
	}
	Logg.Debugf("end new VM run", "code", code)

	v, err := en.st.MatchFlag(state.FLAG_TERMINATE, false)
	if err != nil {
		return false, err
	}
	if v {
		if len(code) > 0 {
			Logg.Debugf("terminated with code remaining", "code", code)
		}
		return false, err
	}

	en.st.SetCode(code)
	if len(code) == 0 {
		Logg.Infof("runner finished with no remaining code")
		_, err = en.reset(ctx)
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

// start execution over at top node while keeping current state of client error flags.
func(en *Engine) reset(ctx context.Context) (bool, error) {
	var err error
	var isTop bool
	for !isTop {
		isTop, err = en.st.Top()
		if err != nil {
			return false, err
		}
		_, err = en.st.Up()
		if err != nil {
			return false, err
		}
		en.ca.Pop()
	}
	en.st.Restart()
	en.initd = false
	return en.Init(ctx)
}
