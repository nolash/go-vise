package vm

import (
	"context"
	"fmt"
	"log"

	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/render"
	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
)

type Vm struct {
	st *state.State
	rs resource.Resource
	pg *render.Page
	ca cache.Memory
	mn *render.Menu
	sizer *render.Sizer
}


func NewVm(st *state.State, rs resource.Resource, ca cache.Memory, sizer *render.Sizer) *Vm {
	vmi := &Vm{
		st: st,
		rs: rs,
		ca: ca,
		pg: render.NewPage(ca, rs),
		sizer: sizer,
	}
	vmi.Reset()
	return vmi
}

func(vmi *Vm) Reset() {
	vmi.mn = render.NewMenu()
	vmi.pg.Reset()
	vmi.pg = vmi.pg.WithMenu(vmi.mn) //render.NewPage(vmi.ca, vmi.rs).WithMenu(vmi.mn)
	if vmi.sizer != nil {
		vmi.pg = vmi.pg.WithSizer(vmi.sizer)	
	}
}

//type Runner func(instruction []byte, st state.State, rs resource.Resource, ctx context.Context) (state.State, []byte, error)

// Run extracts individual op codes and arguments and executes them.
//
// Each step may update the state.
//
// On error, the remaining instructions will be returned. State will not be rolled back.
func(vm *Vm) Run(b []byte, ctx context.Context) ([]byte, error) {
	running := true
	for running {
		r, err := vm.st.MatchFlag(state.FLAG_TERMINATE, false)
		if err != nil {
			panic(err)
		}
		if r {
			log.Printf("terminate set! bailing!")
			return []byte{}, nil
		}
		_, err = vm.st.SetFlag(state.FLAG_DIRTY)
		if err != nil {
			panic(err)
		}
		op, bb, err := opSplit(b)
		if err != nil {
			return b, err
		}
		b = bb
		log.Printf("execute code %x (%s) %x", op, OpcodeString[op], b)
		log.Printf("state: %v", vm.st)
		switch op {
		case CATCH:
			b, err = vm.RunCatch(b, ctx)
		case CROAK:
			b, err = vm.RunCroak(b, ctx)
		case LOAD:
			b, err = vm.RunLoad(b, ctx)
		case RELOAD:
			b, err = vm.RunReload(b, ctx)
		case MAP:
			b, err = vm.RunMap(b, ctx)
		case MOVE:
			b, err = vm.RunMove(b, ctx)
		case INCMP:
			b, err = vm.RunInCmp(b, ctx)
		case MSIZE:
			b, err = vm.RunMSize(b, ctx)
		case MOUT:
			b, err = vm.RunMOut(b, ctx)
		case MNEXT:
			b, err = vm.RunMNext(b, ctx)
		case MPREV:
			b, err = vm.RunMPrev(b, ctx)
		case HALT:
			b, err = vm.RunHalt(b, ctx)
			return b, err
		default:
			err = fmt.Errorf("Unhandled state: %v", op)
		}
		if err != nil {
			return b, err
		}
		if len(b) == 0 {
			b, err = vm.RunDeadCheck(b, ctx)
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

// RunDeadCheck determines whether a state of empty bytecode should result in termination.
//
// If there is remaining bytecode, this method is a noop.
//
// If input has not been matched, a default invalid input page should be generated aswell as a possiblity of return to last screen (or exit).
// 
// If the termination flag has been set but not yet handled, execution is allowed to terminate.
func(vm *Vm) RunDeadCheck(b []byte, ctx context.Context) ([]byte, error) {
	if len(b) > 0 {
		return b, nil
	}
	r, err := vm.st.MatchFlag(state.FLAG_READIN, true)
	if err != nil {
		panic(err)
	}
	if r {
		log.Printf("Not processing input. Setting terminate")
		_, err := vm.st.SetFlag(state.FLAG_TERMINATE)
		if err != nil {
			panic(err)
		}
		return b, nil
	}
	r, err = vm.st.MatchFlag(state.FLAG_TERMINATE, false)
	if err != nil {
		panic(err)
	}
	if r {
		log.Printf("Terminate found!!")
		return b, nil
	}


	log.Printf("no code remaining but not terminating")
	location, _ := vm.st.Where()
	if location == "" {
		return b, fmt.Errorf("dead runner with no current location")
	}
	b = NewLine(nil, MOVE, []string{"_catch"}, nil, nil)
	return b, nil
}

// RunMap executes the MAP opcode
func(vm *Vm) RunMap(b []byte, ctx context.Context) ([]byte, error) {
	sym, b, err := ParseMap(b)
	err = vm.pg.Map(sym)
	return b, err
}

// RunMap executes the CATCH opcode
func(vm *Vm) RunCatch(b []byte, ctx context.Context) ([]byte, error) {
	sym, sig, mode, b, err := ParseCatch(b)
	if err != nil {
		return b, err
	}
	r, err := vm.st.MatchFlag(sig, mode)
	if err != nil {
		return b, err
	}
	if r {
		log.Printf("catch at flag %v, moving to %v", sig, sym) //bitField, d)
		vm.st.Down(sym)
		vm.Reset()
		b = []byte{}
	} 
	return b, nil
}

// RunMap executes the CROAK opcode
func(vm *Vm) RunCroak(b []byte, ctx context.Context) ([]byte, error) {
	sig, mode, b, err := ParseCroak(b)
	if err != nil {
		return b, err
	}
	r, err := vm.st.MatchFlag(sig, mode)
	if err != nil {
		return b, err
	}
	if r {
		log.Printf("croak at flag %v, purging and moving to top", sig)
		vm.Reset()
		vm.st.Reset()
		vm.pg.Reset()
		vm.ca.Reset()
		b = []byte{}
	}
	return []byte{}, nil
}

// RunLoad executes the LOAD opcode
func(vm *Vm) RunLoad(b []byte, ctx context.Context) ([]byte, error) {
	sym, sz, b, err := ParseLoad(b)
	if err != nil {
		return b, err
	}
	r, err := refresh(sym, vm.rs, ctx)
	if err != nil {
		return b, err
	}
	err = vm.ca.Add(sym, r, uint16(sz))
	return b, err
}

// RunLoad executes the RELOAD opcode
func(vm *Vm) RunReload(b []byte, ctx context.Context) ([]byte, error) {
	sym, b, err := ParseReload(b)
	if err != nil {
		return b, err
	}

	r, err := refresh(sym, vm.rs, ctx)
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

// RunLoad executes the MOVE opcode
func(vm *Vm) RunMove(b []byte, ctx context.Context) ([]byte, error) {
	sym, b, err := ParseMove(b)
	if err != nil {
		return b, err
	}
	if sym == "_" {
		vm.st.Up()
		vm.ca.Pop()
		sym, _ = vm.st.Where()
	} else {
		vm.st.Down(sym)
		vm.ca.Push()
	}
	code, err := vm.rs.GetCode(sym)
	if err != nil {
		return b, err
	}
	log.Printf("loaded additional code: %x", code)
	b = append(b, code...)
	vm.Reset()
	return b, nil
}

// RunIncmp executes the INCMP opcode
func(vm *Vm) RunInCmp(b []byte, ctx context.Context) ([]byte, error) {
	sym, target, b, err := ParseInCmp(b)
	if err != nil {
		return b, err
	}

	change, err := vm.st.SetFlag(state.FLAG_READIN)
	if err != nil {
		panic(err)
	}
	have, err := vm.st.GetFlag(state.FLAG_INMATCH)
	if err != nil {
		panic(err)
	}
	if have {
		if change {
			_, err = vm.st.ResetFlag(state.FLAG_INMATCH)
			if err != nil {
				panic(err)
			}
		} else {
			return b, nil
		}
	}
	input, err := vm.st.GetInput()
	if err != nil {
		return b, err
	}
	log.Printf("sym is %s", sym)
	if sym == "*" {
		log.Printf("input wildcard match ('%s'), target '%s'", input, target)
	} else {
		if sym != string(input) {
			return b, nil
		}
		log.Printf("input match for '%s', target '%s'", input, target)
	}

	_, err = vm.st.SetFlag(state.FLAG_INMATCH)
	if err != nil {
		panic(err)
	}
	_, err = vm.st.ResetFlag(state.FLAG_READIN)
	if err != nil {
		panic(err)
	}

	target, _, err = applyTarget([]byte(target), vm.st, vm.ca, ctx)
	if err != nil {
		return b, err
	}
	vm.Reset()

	code, err := vm.rs.GetCode(target)
	if err != nil {
		return b, err
	}
	log.Printf("loaded additional code for target '%s': %x", target, code)
	b = append(b, code...)
	return b, err
}

// RunHalt executes the HALT opcode
func(vm *Vm) RunHalt(b []byte, ctx context.Context) ([]byte, error) {
	var err error
	b, err = ParseHalt(b)
	if err != nil {
		return b, err
	}
	log.Printf("found HALT, stopping")
	return b, err
}

// RunMSize executes the MSIZE opcode
func(vm *Vm) RunMSize(b []byte, ctx context.Context) ([]byte, error) {
	log.Printf("WARNING MSIZE not yet implemented")
	_, _, b, err := ParseMSize(b)
	return b, err
}

// RunMOut executes the MOUT opcode
func(vm *Vm) RunMOut(b []byte, ctx context.Context) ([]byte, error) {
	choice, title, b, err := ParseMOut(b)
	if err != nil {
		return b, err
	}
	err = vm.mn.Put(choice, title)
	return b, err
}

// RunMNext executes the MNEXT opcode
func(vm *Vm) RunMNext(b []byte, ctx context.Context) ([]byte, error) {
       selector, display, b, err := ParseMNext(b)
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
	
// RunMPrev executes the MPREV opcode
func(vm *Vm) RunMPrev(b []byte, ctx context.Context) ([]byte, error) {
       selector, display, b, err := ParseMPrev(b)
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

func(vm *Vm) Render(ctx context.Context) (string, error) {
	changed, err := vm.st.ResetFlag(state.FLAG_DIRTY)
	if err != nil {
		panic(err)	
	}
	if !changed {
		log.Printf("Render called when not dirty, please investigate.")
	}
	sym, idx := vm.st.Where()
	r, err := vm.pg.Render(sym, idx)
	var ok bool
	_, ok = err.(*render.BrowseError)
	if ok {
		vm.Reset()
		b := NewLine(nil, MOVE, []string{"_catch"}, nil, nil)
		vm.Run(b, ctx)
		sym, idx := vm.st.Where()
		r, err = vm.pg.Render(sym, idx)
	}
	if err != nil {
		return "", err
	}
	return r, nil
}

// retrieve data for key
func refresh(key string, rs resource.Resource, ctx context.Context) (string, error) {
	fn, err := rs.FuncFor(key)
	if err != nil {
		return "", err
	}
	if fn == nil {
		return "", fmt.Errorf("no retrieve function for external symbol %v", key)
	}
	return fn(key, ctx)
}

