package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/render"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/vm"
)

var (
	ErrFlushNoExec = errors.New("Attempted flush on unexecuted engine")
)

type DefaultEngine struct {
	st         *state.State
	ca         cache.Memory
	vm         *vm.Vm
	rs         resource.Resource
	pe         *persist.Persister
	cfg        Config
	dbg        Debug
	first      resource.EntryFunc
	initd      bool
	exit       string
	exiting    bool
	execd      bool
	regexCount int
}

// NewEngine instantiates the default Engine implementation.
func NewEngine(cfg Config, rs resource.Resource) *DefaultEngine {
	if rs == nil {
		panic("resource cannot be nil")
	}
	en := &DefaultEngine{
		rs:  rs,
		cfg: cfg,
	}
	if en.cfg.Root == "" {
		en.cfg.Root = "root"
	}
	return en
}

// WithState is a chainable method that explicitly sets the state object to use for the engine.
//
// If not set, the state.State object provided by the persist.Persister will be used.
// If that is not available, a new instance will be created according to engine.Config.
//
// Note that engine.Init will fail if state is set both explicitly
// and in a provided persist.Persister.
func (en *DefaultEngine) WithState(st *state.State) *DefaultEngine {
	if en.st != nil {
		panic("state already set")
	}
	if st == nil {
		panic("state argument is nil")
	}
	en.st = st
	return en
}

// WithMemory is a chainable method that explicitly sets the memory object to use for the engine.
//
// If not set, the cache.Memory object provided by the persist.Persister will be used.
// If that is not available, a new instance will be created according to engine.Config.
//
// Note that engine.Init will fail if memory is set both explicitly
// and in a provided persist.Persister.
func (en *DefaultEngine) WithMemory(ca cache.Memory) *DefaultEngine {
	if en.ca != nil {
		panic("cache already set")
	}
	if ca == nil {
		panic("cache argument is nil")
	}
	en.ca = ca
	return en
}

// WithPersister is a chainable method that sets the persister to use with the engine.
//
// If the persister is missing state, memory or both, it will inherit them from the engine.
func (en *DefaultEngine) WithPersister(pe *persist.Persister) *DefaultEngine {
	if en.pe != nil {
		panic("persister already set")
	}
	if pe == nil {
		panic("persister argument is nil")
	}
	en.pe = pe
	return en
}

// WithDebug is a chainable method that sets the debugger to use for the engine.
//
// If the argument is nil, the default debugger will be used.
func (en *DefaultEngine) WithDebug(dbg Debug) *DefaultEngine {
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

// WithFirst is a chainable method that defines the function that will be run before
// control is handed over to the VM bytecode from the current state.
//
// If this function returns an error, execution will be aborted in engine.Init.
func (en *DefaultEngine) WithFirst(fn resource.EntryFunc) *DefaultEngine {
	if en.first != nil {
		panic("firstfunc already set")
	}
	if fn == nil {
		panic("firstfunc argument is nil")
	}
	en.first = fn
	return en
}

// AddValidInput defines a regular expressing string to match input against.
//
// The added regular expression will be evaluated after the builtin match (see
// vm/input.go for the actual string details).
//
// The function may be called more than once. Input will be validated against each
// in the sequence they were added.
//
// When a match is found, remaining regular expressions will be skipped.
func (en *DefaultEngine) AddValidInput(re string) error {
	err := vm.RegisterInputValidator(en.regexCount, re)
	en.regexCount += 1
	return err
}

// ensure state is present in engine.
func (en *DefaultEngine) ensureState() {
	if en.st == nil {
		st := state.NewState(en.cfg.FlagCount)
		en.st = st
		en.st.SetLanguage(en.cfg.Language)
		if en.st.Language != nil {
			en.st.SetFlag(state.FLAG_LANG)
		}
		logg.Debugf("new engine state added", "state", en.st)
	} else {
		if en.cfg.Language != "" {
			if en.st.Language == nil {
				en.st.SetLanguage(en.cfg.Language)
				en.st.SetFlag(state.FLAG_LANG)
			} else {
				logg.Warnf("language '%s'set in config, but will be ignored because state language has already been set.")
			}
		}
	}
}

// ensure memory is present in engine.
func (en *DefaultEngine) ensureMemory() error {
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

// retrieve state and memory from perister if present.
func (en *DefaultEngine) preparePersist() error {
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
		}
	}
	return nil
}

// synchronize state and memory between engine and persister.
func (en *DefaultEngine) ensurePersist() error {
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
	en.pe = en.pe.WithContent(st, cac)
	err := en.pe.Load(en.cfg.SessionId)
	if err != nil {
		logg.Infof("persister load fail. trying save in case new session", "err", err, "session", en.cfg.SessionId)
		err = en.pe.Save(en.cfg.SessionId)
		if err != nil {
			return err
		}
		en.pe = en.pe.WithContent(st, cac)
		err = en.pe.Load(en.cfg.SessionId)
	}
	if en.cfg.StateDebug {
		en.st.UseDebug()
	}
	logg.Tracef("set persister", "st", st, "cac", cac, "session", en.cfg.SessionId, "persister", en.pe)
	return err
}

// create vm instance.
func (en *DefaultEngine) setupVm() {
	var szr *render.Sizer
	if en.cfg.OutputSize > 0 {
		szr = render.NewSizer(en.cfg.OutputSize)
	}
	en.vm = vm.NewVm(en.st, en.rs, en.ca, szr)
	if en.cfg.MenuSeparator != "" {
		en.vm = en.vm.WithMenuSeparator(en.cfg.MenuSeparator)
	}
}

func (en *DefaultEngine) empty(ctx context.Context) error {
	var err error
	b := bytes.NewBuffer(nil)
	_, err = en.Flush(ctx, b)
	if err != nil {
		return err
	}
	logg.DebugCtxf(ctx, "discard", "output", b.Bytes())
	return nil
}

// prepare engine for Init run.
func (en *DefaultEngine) prepare(ctx context.Context) error {
	if en.execd {
		err := en.empty(ctx)
		if err != nil {
			return err
		}
	}
	en.execd = false
	en.exit = ""
	en.exiting = false
	if en.initd {
		return nil
	}
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

// execute the first function, if set.
func (en *DefaultEngine) runFirst(ctx context.Context) (bool, error) {
	var err error
	var r bool
	if en.first == nil {
		return true, nil
	}
	logg.DebugCtxf(ctx, "start pre-VM check")
	en.ca.Push()
	rs := resource.NewMenuResource()
	rs.AddLocalFunc("_first", en.first)
	en.st.Down("_first")
	defer en.ca.Pop()
	defer en.st.Up()
	defer en.st.ResetFlag(state.FLAG_TERMINATE)
	defer en.st.ResetFlag(state.FLAG_DIRTY)
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
			en.execd = true
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
	logg.DebugCtxf(ctx, "end pre-VM check")
	return r, err
}

// Finish implements the Engine interface.
//
// If persister is set, this call will save the state and memory.
//
// An error will be logged and returned if:
//   - persistence was attempted and failed (takes precedence)
//   - resource backend did not close cleanly.
func (en *DefaultEngine) Finish(ctx context.Context) error {
	var perr error
	if !en.initd {
		return nil
	}
	if en.pe != nil {
		perr = en.pe.Save(en.cfg.SessionId)
	}
	err := en.rs.Close(ctx)
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

func (en *DefaultEngine) setCode(ctx context.Context, code []byte) (bool, error) {
	var err error

	cont := true
	en.st.SetCode(code)
	if len(code) == 0 {
		logg.InfoCtxf(ctx, "runner finished with no remaining code", "state", en.st)
		if en.st.MatchFlag(state.FLAG_DIRTY, true) {
			logg.Debugf("have output for quitting")
			en.exiting = true
			en.exit = en.ca.Last()
		}
		cont = false
	}
	return cont, err
}

// Init implements the Engine interface.
//
// It loads and executes code for the start node.
func (en *DefaultEngine) init(ctx context.Context, input []byte) (bool, error) {
	cont := true
	err := en.prepare(ctx)
	if err != nil {
		return false, err
	}

	if en.st.Language != nil {
		logg.TraceCtxf(ctx, "set language on context", "lang", en.st.Language)
		ctx = context.WithValue(ctx, "Language", *en.st.Language)
	}

	if en.initd {
		logg.DebugCtxf(ctx, "already initialized")
		return true, nil
	}

	sym := en.cfg.Root
	if sym == "" {
		return false, fmt.Errorf("start sym empty")
	}

	inSave, _ := en.st.GetInput()
	err = en.st.SetInput(input)
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

	if len(en.st.Code) == 0 {
		b := vm.NewLine(nil, vm.MOVE, []string{sym}, nil, nil)
		cont, err = en.setCode(ctx, b)
		if err != nil {
			return false, err
		}
	}

	err = en.st.SetInput(inSave)
	if err != nil {
		return false, err
	}
	en.initd = true
	return cont, nil
}

// Exec implements the Engine interface.
//
// It processes user input against the current state of the virtual machine environment.
//
// If successfully executed, output of the last execution is available using the Flush call.
//
// A bool return valus of false indicates that execution should be terminated. Calling Exec again has undefined effects.
//
// Fails if:
//   - input is formally invalid (too long etc)
//   - no current bytecode is available
//   - input processing against bytcode failed
func (en *DefaultEngine) Exec(ctx context.Context, input []byte) (bool, error) {
	var err error

	if en.cfg.SessionId != "" {
		ctx = context.WithValue(ctx, "SessionId", en.cfg.SessionId)
	}

	cont, err := en.init(ctx, input)
	if err != nil {
		return false, err
	}
	if !cont {
		return cont, nil
	}

	if en.st.Language != nil {
		ctx = context.WithValue(ctx, "Language", *en.st.Language)
	}

	if len(input) > 0 {
		_, err = vm.ValidInput(input)
		if err != nil {
			return true, err
		}
	}
	err = en.st.SetInput(input)
	if err != nil {
		return false, err
	}
	return en.exec(ctx, input)
}

// backend for Exec, after the input validity check
func (en *DefaultEngine) exec(ctx context.Context, input []byte) (bool, error) {
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
	en.execd = true
	logg.Debugf("end new VM run", "code", code)

	v := en.st.MatchFlag(state.FLAG_TERMINATE, true)
	if v {
		if len(code) > 0 {
			logg.Debugf("terminated with code remaining", "code", code)
		}
		return false, err
	}
	cont, err := en.setCode(ctx, code)
	if en.dbg != nil {
		en.dbg.Break(en.st, en.ca)
	}
	return cont, err
}

// Flush implements the Engine interface.
//
// The method writes the output of the last vm execution to the given writer.
//
// Fails if
//   - required data inputs to the template are not available.
//   - the template for the given node point is note available for retrieval using the resource.Resource implementer.
//   - the supplied writer fails to process the writes.
func (en *DefaultEngine) Flush(ctx context.Context, w io.Writer) (int, error) {
	var l int
	if !en.execd {
		return 0, ErrFlushNoExec
	}
	if en.st.Language != nil {
		ctx = context.WithValue(ctx, "Language", *en.st.Language)
	}
	logg.TraceCtxf(ctx, "render with state", "state", en.st)
	r, err := en.vm.Render(ctx)
	if err != nil {
		if len(en.exit) == 0 {
			return 0, err
		}
	} else {
		if len(r) > 0 {
			l, err = io.WriteString(w, r)
			if err != nil {
				return l, err
			}
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
	if en.exiting {
		_, err = en.reset(ctx)
		en.exiting = false
	}

	return l, err
}

// start execution over at top node while keeping current state of client error flags.
func (en *DefaultEngine) Reset(ctx context.Context, force bool) (bool, error) {
	var err error
	var b []byte
	if force {
		en.st.SetCode(b)
	} else {
		b, err = en.st.GetCode()
		if err != nil {
			return false, err
		}
		if len(b) > 0 {
			return false, ErrCodeRemaining
		}
	}
	return en.reset(ctx)
}

func (en *DefaultEngine) reset(ctx context.Context) (bool, error) {
	var err error
	var isTop bool
	logg.DebugCtxf(ctx, "entering engine reset", "state", en.st)
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
	en.st.ResetFlag(state.FLAG_TERMINATE)
	en.st.ResetFlag(state.FLAG_DIRTY)
	return false, nil
}
