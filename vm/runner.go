package vm

import (
	"context"
	"errors"
	"fmt"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/render"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
)

// ExternalCodeError indicates an error that occurred when resolving an external code symbol (LOAD, RELOAD).
type ExternalCodeError struct {
	sym string
	code int
	err error
}

// NewExternalCodeError creates a new ExternalCodeError.
func NewExternalCodeError(sym string, err error) *ExternalCodeError {
	return &ExternalCodeError{
		sym: sym,
		err: err,
	}
}

func(e *ExternalCodeError) WithCode(code int) *ExternalCodeError {
	e.code = code
	return e
}

// Error implements the Error interface.
func(e ExternalCodeError) Error() string {
	logg.Errorf("external code error: %v", e.err)
	return fmt.Sprintf("error %v:%v", e.sym, e.code)
}

// Vm holds sub-components mutated by the vm execution.
// TODO: Renderer should be passed to avoid proxy methods not strictly related to vm operation
type Vm struct {
	st *state.State // Navigation and error states.
	rs resource.Resource // Retrieves content, code, and templates for symbols.
	ca cache.Memory // Loaded content.
	mn *render.Menu // Menu component of page.
	sizer *render.Sizer // Apply size constraints to output.
	pg *render.Page // Render outputs with menues to size constraints.
}

// NewVm creates a new Vm.
func NewVm(st *state.State, rs resource.Resource, ca cache.Memory, sizer *render.Sizer) *Vm {
	vmi := &Vm{
		st: st,
		rs: rs,
		ca: ca,
		pg: render.NewPage(ca, rs),
		sizer: sizer,
	}
	vmi.Reset()
	logg.Infof("vm created with state", "state", st, "renderer", vmi.pg)
	return vmi
}

// Reset re-initializes sub-components for output rendering.
func(vmi *Vm) Reset() {
	vmi.mn = render.NewMenu()
	vmi.pg.Reset()
	vmi.pg = vmi.pg.WithMenu(vmi.mn)
	if vmi.sizer != nil {
		vmi.pg = vmi.pg.WithSizer(vmi.sizer)	
	}
}

// Run extracts individual op codes and arguments and executes them.
//
// Each step may update the state.
//
// On error, the remaining instructions will be returned. State will not be rolled back.
func(vm *Vm) Run(ctx context.Context, b []byte) ([]byte, error) {
	logg.Tracef("new vm run")
	running := true
	for running {
		r := vm.st.MatchFlag(state.FLAG_TERMINATE, true)
		if r {
			logg.InfoCtxf(ctx, "terminate set! bailing")
			return []byte{}, nil
		}

		_ = vm.st.ResetFlag(state.FLAG_TERMINATE)

		change := vm.st.ResetFlag(state.FLAG_LANG)
		if change {
			if vm.st.Language != nil {
				ctx = context.WithValue(ctx, "Language", *vm.st.Language)
			}
		}

		waitChange := vm.st.ResetFlag(state.FLAG_WAIT)
		if waitChange {
			vm.st.ResetFlag(state.FLAG_INMATCH)
			vm.pg.Reset()
			vm.mn.Reset()
		}

		_ = vm.st.SetFlag(state.FLAG_DIRTY)
		op, bb, err := opSplit(b)
		if err != nil {
			return b, err
		}
		b = bb
		logg.DebugCtxf(ctx, "execute code", "opcode", op, "op", OpcodeString[op], "code", b)
		logg.DebugCtxf(ctx, "", "state", vm.st)
		switch op {
		case CATCH:
			b, err = vm.runCatch(ctx, b)
		case CROAK:
			b, err = vm.runCroak(ctx, b)
		case LOAD:
			b, err = vm.runLoad(ctx, b)
		case RELOAD:
			b, err = vm.runReload(ctx, b)
		case MAP:
			b, err = vm.runMap(ctx, b)
		case MOVE:
			b, err = vm.runMove(ctx, b)
		case INCMP:
			b, err = vm.runInCmp(ctx, b)
		case MSINK:
			b, err = vm.runMSink(ctx, b)
		case MOUT:
			b, err = vm.runMOut(ctx, b)
		case MNEXT:
			b, err = vm.runMNext(ctx, b)
		case MPREV:
			b, err = vm.runMPrev(ctx, b)
		case HALT:
			b, err = vm.runHalt(ctx, b)
			return b, err
		default:
			err = fmt.Errorf("Unhandled state: %v", op)
		}
		b, err = vm.runErrCheck(ctx, b, err)
		if err != nil {
			return b, err
		}
		if len(b) == 0 {
			b, err = vm.runDeadCheck(ctx, b)
			if err != nil {
				return b, err
			}
		}
		if len(b) == 0 {
			return []byte{}, nil
		}
	}
	return b, nil
}

// handles errors that should not be deferred to the client.
func(vm *Vm) runErrCheck(ctx context.Context, b []byte, err error) ([]byte, error) {
	if err == nil {
		return b, err
	}
	vm.pg = vm.pg.WithError(err)

	v := vm.st.MatchFlag(state.FLAG_LOADFAIL, true)
	if !v {
		return b, err
	}

	b = NewLine(nil, MOVE, []string{"_catch"}, nil, nil)
	return b, nil
}

// determines whether a state of empty bytecode should result in termination.
//
// If there is remaining bytecode, this method is a noop.
//
// If input has not been matched, a default invalid input page should be generated aswell as a possiblity of return to last screen (or exit).
// 
// If the termination flag has been set but not yet handled, execution is allowed to terminate.
func(vm *Vm) runDeadCheck(ctx context.Context, b []byte) ([]byte, error) {
	if len(b) > 0 {
		return b, nil
	}
	r := vm.st.MatchFlag(state.FLAG_READIN, false)
	if r {
		logg.DebugCtxf(ctx, "Not processing input. Setting terminate")
		vm.st.SetFlag(state.FLAG_TERMINATE)
		return b, nil
	}
	r = vm.st.MatchFlag(state.FLAG_TERMINATE, true)
	if r {
		logg.TraceCtxf(ctx, "Terminate found!!")
		return b, nil
	}

	logg.TraceCtxf(ctx, "no code remaining but not terminating")
	location, _ := vm.st.Where()
	if location == "" {
		return b, fmt.Errorf("dead runner with no current location")
	} else if location == "_catch" {
		return b, fmt.Errorf("unexpected catch endless loop detected for state: %s", vm.st)
	}

	input, err := vm.st.GetInput()
	if err != nil {
		input = []byte("(no input)")
	}
	cerr := NewInvalidInputError(string(input))
	vm.pg.WithError(cerr)
	b = NewLine(nil, MOVE, []string{"_catch"}, nil, nil)
	return b, nil
}

// executes the MAP opcode
func(vm *Vm) runMap(ctx context.Context, b []byte) ([]byte, error) {
	sym, b, err := ParseMap(b)
	err = vm.pg.Map(sym)
	return b, err
}

// executes the CATCH opcode
func(vm *Vm) runCatch(ctx context.Context, b []byte) ([]byte, error) {
	sym, sig, mode, b, err := ParseCatch(b)
	if err != nil {
		return b, err
	}
	r := vm.st.MatchFlag(sig, mode)
	if r {
		actualSym, _, err := applyTarget([]byte(sym), vm.st, vm.ca, ctx)
		if err != nil {
			return b, err
		}
		logg.InfoCtxf(ctx, "catch!", "flag", sig, "sym", sym, "target", actualSym, "mode", mode)
		sym = actualSym
		bh, err := vm.rs.GetCode(ctx, sym)
		if err != nil {
			return b, err
		}
		b = bh
	}
	return b, nil
}

// executes the CROAK opcode
func(vm *Vm) runCroak(ctx context.Context, b []byte) ([]byte, error) {
	sig, mode, b, err := ParseCroak(b)
	if err != nil {
		return b, err
	}
	r := vm.st.MatchFlag(sig, mode)
	if r {
		logg.InfoCtxf(ctx, "croak! purging and moving to top", "signal", sig)
		vm.Reset()
		vm.ca.Reset()
		b = []byte{}
	}
	return b, nil
}

// executes the LOAD opcode
func(vm *Vm) runLoad(ctx context.Context, b []byte) ([]byte, error) {
	sym, sz, b, err := ParseLoad(b)
	if err != nil {
		return b, err
	}
	_, err = vm.ca.Get(sym)
	if err == nil {
		logg.DebugCtxf(ctx, "skip already loaded symbol", "symbol", sym)
		return b, nil
	}
	r, err := vm.refresh(sym, vm.rs, ctx)
	if err != nil {
		return b, err
	}
	err = vm.ca.Add(sym, r, uint16(sz))
	if err != nil {
		if err == cache.ErrDup {
			logg.DebugCtxf(ctx, "Ignoring load request on frame that has symbol already loaded", "sym", sym)
			err = nil	
		}
	}
	return b, err
}

// executes the RELOAD opcode
func(vm *Vm) runReload(ctx context.Context, b []byte) ([]byte, error) {
	sym, b, err := ParseReload(b)
	if err != nil {
		return b, err
	}

	r, err := vm.refresh(sym, vm.rs, ctx)
	if err != nil {
		return b, err
	}
	vm.ca.Update(sym, r)
	if vm.pg != nil {
		err := vm.pg.Map(sym)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}

// executes the MOVE opcode
func(vm *Vm) runMove(ctx context.Context, b []byte) ([]byte, error) {
	sym, b, err := ParseMove(b)
	if err != nil {
		return b, err
	}
	sym, _, err = applyTarget([]byte(sym), vm.st, vm.ca, ctx)
	if err != nil {
		return b, err
	}
	code, err := vm.rs.GetCode(ctx, sym)
	if err != nil {
		return b, err
	}
	logg.DebugCtxf(ctx, "loaded code", "sym", sym, "code", code)
	b = append(b, code...)
	vm.Reset()
	return b, nil
}

// executes the INCMP opcode
// TODO: document state transition table and simplify flow
func(vm *Vm) runInCmp(ctx context.Context, b []byte) ([]byte, error) {
	sym, target, b, err := ParseInCmp(b)
	if err != nil {
		return b, err
	}

	reading := vm.st.GetFlag(state.FLAG_READIN)
	have := vm.st.GetFlag(state.FLAG_INMATCH)
	if err != nil {
		panic(err)
	}
	if have {
		if reading {
			logg.DebugCtxf(ctx, "ignoring input - already have match", "input", sym)
			return b, nil
		}
	} else {
		vm.st.SetFlag(state.FLAG_READIN)
	}
	input, err := vm.st.GetInput()
	if err != nil {
		return b, err
	}
	logg.TraceCtxf(ctx, "testing sym", "sym", sym, "input", input)

	if !have && target == "*" {
		logg.DebugCtxf(ctx, "input wildcard match", "input", input, "next", sym)
	} else {
		if target != string(input) {
			return b, nil
		} 
		logg.InfoCtxf(ctx, "input match", "input", input, "next", sym)
	}
	vm.st.SetFlag(state.FLAG_INMATCH)
	vm.st.ResetFlag(state.FLAG_READIN)

	newSym, _, err := applyTarget([]byte(sym), vm.st, vm.ca, ctx)

	//_, ok := err.(*state.IndexError)
	//if ok {
	if errors.Is(err, state.IndexError) {
		vm.st.SetFlag(state.FLAG_READIN)
		return b, nil
	} else if err != nil {
		return b, err
	}

	sym = newSym

	vm.Reset()

	code, err := vm.rs.GetCode(ctx, sym)
	if err != nil {
		return b, err
	}
	logg.DebugCtxf(ctx, "loaded additional code", "next", sym, "code", code)
	b = append(b, code...)
	return b, err
}

// executes the HALT opcode
func(vm *Vm) runHalt(ctx context.Context, b []byte) ([]byte, error) {
	var err error
	b, err = ParseHalt(b)
	if err != nil {
		return b, err
	}
	logg.DebugCtxf(ctx, "found HALT, stopping")
	
	vm.st.SetFlag(state.FLAG_WAIT)
	return b, nil
}

// executes the MSIZE opcode
func(vm *Vm) runMSink(ctx context.Context, b []byte) ([]byte, error) {
	b, err := ParseMSink(b)
	mcfg := vm.mn.GetBrowseConfig()
	vm.mn = vm.mn.WithSink().WithBrowseConfig(mcfg).WithPages()
	//vm.pg.WithMenu(vm.mn)
	return b, err
}

// executes the MOUT opcode
func(vm *Vm) runMOut(ctx context.Context, b []byte) ([]byte, error) {
	title, choice, b, err := ParseMOut(b)
	if err != nil {
		return b, err
	}
	err = vm.mn.Put(choice, title)
	return b, err
}

// executes the MNEXT opcode
func(vm *Vm) runMNext(ctx context.Context, b []byte) ([]byte, error) {
       display, selector, b, err := ParseMNext(b)
       if err != nil {
	       return b, err
       }
       cfg := vm.mn.GetBrowseConfig()
       cfg.NextSelector = selector
       cfg.NextTitle = display
       cfg.NextAvailable = true
       vm.mn = vm.mn.WithBrowseConfig(cfg)
       return b, nil
}
	
// executes the MPREV opcode
func(vm *Vm) runMPrev(ctx context.Context, b []byte) ([]byte, error) {
       display, selector, b, err := ParseMPrev(b)
       if err != nil {
	       return b, err
       }
       cfg := vm.mn.GetBrowseConfig()
       cfg.PreviousSelector = selector
       cfg.PreviousTitle = display
       cfg.PreviousAvailable = true
       vm.mn = vm.mn.WithBrowseConfig(cfg)
       return b, nil
}

// Render wraps output rendering, and handles error when attempting to browse beyond the rendered page count.
func(vm *Vm) Render(ctx context.Context) (string, error) {
	changed := vm.st.ResetFlag(state.FLAG_DIRTY)
	if !changed {
		return "", nil
	}
	sym, idx := vm.st.Where()
	if sym == "" {
		return "", nil
	}
	r, err := vm.pg.Render(ctx, sym, idx)
	var ok bool
	_, ok = err.(*render.BrowseError)
	if ok {
		vm.Reset()
		b := NewLine(nil, MOVE, []string{"_catch"}, nil, nil)
		vm.Run(ctx, b)
		sym, idx := vm.st.Where()
		r, err = vm.pg.Render(ctx, sym, idx)
	}
	if err != nil {
		return "", err
	}
	return r, nil
}

// retrieve and cache data for key
func(vm *Vm) refresh(key string, rs resource.Resource, ctx context.Context) (string, error) {
	var err error
	
	fn, err := rs.FuncFor(ctx, key)
	if err != nil {
		return "", err
	}
	if fn == nil {
		return "", fmt.Errorf("no retrieve function for external symbol %v", key)
	}
	input, _ := vm.st.GetInput()
	r, err := fn(ctx, key, input)
	if err != nil {
		logg.Errorf("external function load fail", "key", key, "error", err)
		_ = vm.st.SetFlag(state.FLAG_LOADFAIL)
		return "", NewExternalCodeError(key, err).WithCode(r.Status)
	}
	for _, flag := range r.FlagReset {
		if !state.IsWriteableFlag(flag) {
			continue
		}
		vm.st.ResetFlag(flag)
	}
	for _, flag := range r.FlagSet {
		if !state.IsWriteableFlag(flag) {
			continue
		}
		vm.st.SetFlag(flag)
	}

	haveLang := vm.st.MatchFlag(state.FLAG_LANG, true)
	if haveLang {
		vm.st.SetLanguage(r.Content)
	}

	return r.Content, err
}
