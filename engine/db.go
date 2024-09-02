package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/render"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/vm"
)

type dbEngine struct {
	st *state.State
	ca cache.Memory
	vm *vm.Vm
	rs resource.Resource
	pe *persist.Persister
	cfg Config
	dbg Debug
	first resource.EntryFunc
	initd bool
	exit string
}

func NewEngine(cfg Config, rs resource.Resource) *dbEngine {
	if rs == nil {
		panic("resource cannot be nil")	
	}
	en := &dbEngine{
		rs: rs,
		cfg: cfg,
	}
	if en.cfg.Root == "" {
		en.cfg.Root = "root"
	}
	return en
}

func(en *dbEngine) WithState(st *state.State) *dbEngine {
	if en.st != nil {
		panic("state already set")
	}
	if st == nil {
		panic("state argument is nil")
	}
	en.st = st
	return en
}

func(en *dbEngine) WithMemory(ca cache.Memory) *dbEngine {
	if en.ca != nil {
		panic("cache already set")
	}
	if ca == nil {
		panic("cache argument is nil")
	}
	en.ca = ca
	return en
}

func(en *dbEngine) WithResource(rs resource.Resource) *dbEngine {
	if en.rs != nil {
		panic("resource already set")
	}
	if rs == nil {
		panic("resource argument is nil")
	}
	en.rs = rs
	return en
}

func(en *dbEngine) WithPersister(pe *persist.Persister) *dbEngine {
	if en.pe != nil {
		panic("persister already set")
	}
	if pe == nil {
		panic("persister argument is nil")
	}
	en.pe = pe 
	return en
}

func(en *dbEngine) WithDebug(dbg Debug) *dbEngine {
	if en.dbg != nil {
		panic("debugger already set")
	}
	if dbg == nil {
		logg.Infof("debug argument was nil, using default debugger")
		dbg = NewSimpleDebug(os.Stderr)
	}
	en.dbg = dbg
	return en
}

func(en *dbEngine) WithFirst(fn resource.EntryFunc) *dbEngine {
	if en.first != nil {
		panic("firstfunc already set")
	}
	if fn == nil {
		panic("firstfunc argument is nil")
	}
	en.first = fn
	return en
}

func(en *dbEngine) ensureState() {
	if en.st == nil {
		st := state.NewState(en.cfg.FlagCount)
		en.st = &st
		en.st.SetLanguage(en.cfg.Language)
		if en.st.Language != nil {
			en.st.SetFlag(state.FLAG_LANG)
		}
		logg.Debugf("new engine state added", "state", en.st)
	} else {
		if (en.cfg.Language != "") {
			if en.st.Language == nil {
				en.st.SetLanguage(en.cfg.Language)
				en.st.SetFlag(state.FLAG_LANG)
			} else {
				logg.Warnf("language '%s'set in config, but will be ignored because state language has already been set.", )
			}
		}
	}
}

func(en *dbEngine) ensureMemory() error {
	cac, ok := en.ca.(*cache.Cache)
	if cac == nil {
		ca := cache.NewCache()
		if en.cfg.CacheSize > 0 {
			ca = ca.WithCacheSize(en.cfg.CacheSize)
		}
		en.ca = ca
		logg.Debugf("new engine memory added", "memory", en.ca)
	} else if !ok {
		return errors.New("memory MUST be *cache.Cache for now. sorry")
	}

	return nil
}

func(en *dbEngine) preparePersist() error {
	if en.pe == nil {
		return nil
	}
	st := en.pe.GetState()
	if st != nil {
		if en.st != nil {
			return errors.New("state cannot be explicitly set in both persister and engine.")
		}
		en.st = st
	} else {
		if en.st == nil {
			logg.Debugf("defer persist state set until state set in engine")
		} else {
			en.st = st
		}
	}

	ca := en.pe.GetMemory()
	cac, ok := ca.(*cache.Cache)
	if !ok {
		return errors.New("memory MUST be *cache.Cache for now. sorry")
	}
	if cac != nil {
		logg.Debugf("ca", "ca", cac)
		if en.ca != nil {
			return errors.New("cache cannot be explicitly set in both persister and engine.")
		}
		en.ca = cac
	} else {
		if en.ca == nil {
			logg.Debugf("defer persist memory set until memory set in engine")
		} else {
			en.ca = cac
		}
	}
	return nil
}

func(en *dbEngine) ensurePersist() error {
	if en.pe == nil {
		return nil
	}
	st := en.pe.GetState()
	if st == nil {
		st = en.st
		logg.Debugf("using engine state for persister", "state", st)
	} else {
		en.st = st
	}
	ca := en.pe.GetMemory()
	cac, ok := ca.(*cache.Cache)
	if cac == nil {
		cac, ok = en.ca.(*cache.Cache)
		if !ok {
			return errors.New("memory MUST be *cache.Cache for now. sorry")
		}
		logg.Debugf("using engine memory for persister", "memory", cac)
	} else {
		en.ca = cac
	}
	logg.Tracef("set persister", "st", st, "cac", cac)
	en.pe = en.pe.WithContent(st, cac)
	err := en.pe.Load(en.cfg.SessionId)
	if err != nil {
		logg.Infof("persister load fail. trying save in case new session", "err", err, "session", en.cfg.SessionId)
		err = en.pe.Save(en.cfg.SessionId)
	}
	if en.cfg.StateDebug {
		en.st.UseDebug()
	}
	return err
}

func(en *dbEngine) setupVm() {
	var szr *render.Sizer
	if en.cfg.OutputSize > 0 {
		szr = render.NewSizer(en.cfg.OutputSize)
	}
	en.vm = vm.NewVm(en.st, en.rs, en.ca, szr)
}

func(en *dbEngine) prepare() error {
	err := en.preparePersist()
	if err != nil {
		return err
	}
	en.ensureState()
	err = en.ensureMemory()
	if err != nil {
		return err
	}
	err = en.ensurePersist()
	if err != nil {
		return err
	}
	en.setupVm()
	return nil
}

// execute the first function, if set
func(en *dbEngine) runFirst(ctx context.Context) (bool, error) {
	var err error
	var r bool
	if en.first == nil {
		return true, nil
	}
	logg.DebugCtxf(ctx, "start pre-VM check")
	rs := resource.NewMenuResource()
	rs.AddLocalFunc("_first", en.first)
	en.st.Down("_first")
	pvm := vm.NewVm(en.st, rs, en.ca, nil)
	b := vm.NewLine(nil, vm.LOAD, []string{"_first"}, []byte{0}, nil)
	b = vm.NewLine(b, vm.HALT, nil, nil, nil)
	b, err = pvm.Run(ctx, b)
	if err != nil {
		return false, err
	}
	if len(b) > 0 {
		// TODO: typed error
		err = fmt.Errorf("Pre-VM code cannot have remaining bytecode after execution, had: %x", b)
	} else {
		if en.st.MatchFlag(state.FLAG_TERMINATE, true) {
			en.exit = en.ca.Last()
			logg.InfoCtxf(ctx, "Pre-VM check says not to continue execution", "state", en.st)
		} else {
			r = true
		}
	}
	if err != nil {
		en.st.Invalidate()
		en.ca.Invalidate()
	}
	en.st.ResetFlag(state.FLAG_TERMINATE)
	en.st.ResetFlag(state.FLAG_DIRTY)
	logg.DebugCtxf(ctx, "end pre-VM check")
	return r, err
}

// Finish implements EngineIsh interface
func(en *dbEngine) Finish() error {
	var perr error
	if en.pe != nil {
		perr = en.pe.Save(en.cfg.SessionId)
	}
	err := en.rs.Close()
	if err != nil {
		logg.Errorf("resource close failed!", "err", err)
	}
	if perr != nil {
		logg.Errorf("persistence failed!", "err", perr)
		err = perr	
	}
	if err == nil {
		logg.Tracef("that's a wrap", "engine", en)
	}
	return err
}

// change root to current state location if non-empty.
func(en *dbEngine) restore() {
	location, _ := en.st.Where()
	if len(location) == 0 {
		return
	}
	if en.cfg.Root != location {
		logg.Infof("restoring state", "sym", location)
		en.cfg.Root = "."
	}
}

// Init must be explicitly called before using the Engine instance.
//
// It loads and executes code for the start node.
func(en *dbEngine) Init(ctx context.Context) (bool, error) {
	err := en.prepare()
	if err != nil {
		return false, err
	}
	en.restore()
	if en.initd {
		logg.DebugCtxf(ctx, "already initialized")
		return true, nil
	}
	
	sym := en.cfg.Root
	if sym == "" {
		return false, fmt.Errorf("start sym empty")
	}

	inSave, _ := en.st.GetInput()
	err = en.st.SetInput([]byte{})
	if err != nil {
		return false, err
	}

	r, err := en.runFirst(ctx)
	if err != nil {
		return false, err
	}
	if !r {
		return false, nil
	}

	b := vm.NewLine(nil, vm.MOVE, []string{sym}, nil, nil)
	logg.DebugCtxf(ctx, "start new init VM run", "code", b)
	b, err = en.vm.Run(ctx, b)
	if err != nil {
		return false, err
	}
	
	logg.DebugCtxf(ctx, "end new init VM run", "code", b)
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
func (en *dbEngine) Exec(ctx context.Context, input []byte) (bool, error) {
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
func(en *dbEngine) exec(ctx context.Context, input []byte) (bool, error) {
	logg.InfoCtxf(ctx, "new VM execution with input", "input", string(input))
	code, err := en.st.GetCode()
	if err != nil {
		return false, err
	}
	if len(code) == 0 {
		return false, fmt.Errorf("no code to execute")
	}

	logg.Debugf("start new VM run", "code", code)
	code, err = en.vm.Run(ctx, code)
	if err != nil {
		return false, err
	}
	logg.Debugf("end new VM run", "code", code)

	v := en.st.MatchFlag(state.FLAG_TERMINATE, true)
	if v {
		if len(code) > 0 {
			logg.Debugf("terminated with code remaining", "code", code)
		}
		return false, err
	}

	en.st.SetCode(code)
	if len(code) == 0 {
		logg.Infof("runner finished with no remaining code", "state", en.st)
		if en.st.MatchFlag(state.FLAG_DIRTY, true) {
			logg.Debugf("have output for quitting")
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
func(en *dbEngine) WriteResult(ctx context.Context, w io.Writer) (int, error) {
	var l int
	if en.st.Language != nil {
		ctx = context.WithValue(ctx, "Language", *en.st.Language)
	}
	logg.TraceCtxf(ctx, "render with state", "state", en.st)
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
		logg.TraceCtxf(ctx, "have exit", "exit", en.exit)
		n, err := io.WriteString(w, en.exit)
		if err != nil {
			return l, err
		}
		l += n
	}
	return l, nil
}

// start execution over at top node while keeping current state of client error flags.
func(en *dbEngine) reset(ctx context.Context) (bool, error) {
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
