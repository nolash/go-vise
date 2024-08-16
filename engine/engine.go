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

// EngineIsh defines the interface for execution engines that handle vm initialization and execution, and rendering outputs.
type EngineIsh interface {
	Init(ctx context.Context) (bool, error)
	Exec(ctx context.Context, input []byte) (bool, error)
	WriteResult(ctx context.Context, w io.Writer) (int, error)
	Finish() error
}

// Config globally defines behavior of all components driven by the engine.
type Config struct {
	OutputSize uint32 // Maximum size of output from a single rendered page
	SessionId string
	Root string
	FlagCount uint32
	CacheSize uint32
	Language string
}

// Engine is an execution engine that handles top-level errors when running client inputs against code in the bytecode buffer.
type Engine struct {
	st *state.State
	rs resource.Resource
	ca cache.Memory
	vm *vm.Vm
	dbg Debug
	root string
	session string
	initd bool
	exit string
}

// NewEngine creates a new Engine
func NewEngine(ctx context.Context, cfg Config, st *state.State, rs resource.Resource, ca cache.Memory) Engine {
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

	var err error
	if st.Language == nil {
		if cfg.Language != "" {
			err = st.SetLanguage(cfg.Language)
			if err != nil {
				panic(err)
			}
			Logg.InfoCtxf(ctx, "set language from config", "language", cfg.Language)
		}
	}

	if st.Language != nil {
		st.SetFlag(state.FLAG_LANG)
	}
	return engine
}

func (en *Engine) SetDebugger(debugger Debug) {
	en.dbg = debugger
}

// Finish implements EngineIsh interface
func(en *Engine) Finish() error {
	Logg.Tracef("that's a wrap", "engine", en)
	return nil
}

// change root to current state location if non-empty.
func(en *Engine) restore() {
	location, _ := en.st.Where()
	if len(location) == 0 {
		return
	}
	if en.root != location {
		Logg.Infof("restoring state", "sym", location)
		en.root = "."
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
	b, err = en.vm.Run(ctx, b)
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
func (en *Engine) Exec(ctx context.Context, input []byte) (bool, error) {
	var err error
	if en.st.Language != nil {
		ctx = context.WithValue(ctx, "Language", *en.st.Language)
	}
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
	return en.exec(ctx, input)
}

// backend for Exec, after the input validity check
func(en *Engine) exec(ctx context.Context, input []byte) (bool, error) {
	Logg.InfoCtxf(ctx, "new VM execution with input", "input", string(input))
	code, err := en.st.GetCode()
	if err != nil {
		return false, err
	}
	if len(code) == 0 {
		return false, fmt.Errorf("no code to execute")
	}

	Logg.Debugf("start new VM run", "code", code)
	code, err = en.vm.Run(ctx, code)
	if err != nil {
		return false, err
	}
	Logg.Debugf("end new VM run", "code", code)

	v := en.st.MatchFlag(state.FLAG_TERMINATE, true)
	if v {
		if len(code) > 0 {
			Logg.Debugf("terminated with code remaining", "code", code)
		}
		return false, err
	}

	en.st.SetCode(code)
	if len(code) == 0 {
		Logg.Infof("runner finished with no remaining code", "state", en.st)
		if en.st.MatchFlag(state.FLAG_DIRTY, true) {
			Logg.Debugf("have output for quitting")
			en.exit = en.ca.Last()
		}
		_, err = en.reset(ctx)
		return false, err
	}

	if en.dbg != nil {
		en.dbg.Break(en.st, en.ca)
	}
	return true, nil
}

// WriteResult writes the output of the last vm execution to the given writer.
//
// Fails if
// - required data inputs to the template are not available.
// - the template for the given node point is note available for retrieval using the resource.Resource implementer.
// - the supplied writer fails to process the writes.
func(en *Engine) WriteResult(ctx context.Context, w io.Writer) (int, error) {
	var l int
	if en.st.Language != nil {
		ctx = context.WithValue(ctx, "Language", *en.st.Language)
	}
	Logg.TraceCtxf(ctx, "render with state", "state", en.st)
	r, err := en.vm.Render(ctx)
	if err != nil {
		return 0, err
	}
	if len(r) > 0 {
		l, err = io.WriteString(w, r)
		if err != nil {
			return l, err
		}
	}
	if len(en.exit) > 0 {
		Logg.TraceCtxf(ctx, "have exit", "exit", en.exit)
		n, err := io.WriteString(w, en.exit)
		if err != nil {
			return l, err
		}
		l += n
	}
	return l, nil
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
	return false, nil
}
