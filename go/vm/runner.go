package vm

import (
	"context"
	"fmt"
	"log"

	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/state"
)

//type Runner func(instruction []byte, st state.State, rs resource.Resource, ctx context.Context) (state.State, []byte, error)

// Run extracts individual op codes and arguments and executes them.
//
// Each step may update the state.
//
// On error, the remaining instructions will be returned. State will not be rolled back.
func Run(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	running := true
	for running {
		log.Printf("execute code %x", b)
		op, bb, err := opSplit(b)
		if err != nil {
			return b, err
		}
		b = bb
		switch op {
		case CATCH:
			b, err = RunCatch(b, st, rs, ctx)
		case CROAK:
			b, err = RunCroak(b, st, rs, ctx)
		case LOAD:
			b, err = RunLoad(b, st, rs, ctx)
		case RELOAD:
			b, err = RunReload(b, st, rs, ctx)
		case MAP:
			b, err = RunMap(b, st, rs, ctx)
		case MOVE:
			b, err = RunMove(b, st, rs, ctx)
		case INCMP:
			b, err = RunInCmp(b, st, rs, ctx)
		case MSIZE:
			b, err = RunMSize(b, st, rs, ctx)
		case MOUT:
			b, err = RunMOut(b, st, rs, ctx)
		case MNEXT:
			b, err = RunMNext(b, st, rs, ctx)
		case MPREV:
			b, err = RunMPrev(b, st, rs, ctx)
		case HALT:
			b, err = RunHalt(b, st, rs, ctx)
			return b, err
		default:
			err = fmt.Errorf("Unhandled state: %v", op)
		}
		if err != nil {
			return b, err
		}
		if len(b) == 0 {
			return []byte{}, nil
		}
	}
	return b, nil
}

// RunMap executes the MAP opcode
func RunMap(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	sym, b, err := ParseMap(b)
	err = st.Map(sym)
	return b, err
}

// RunMap executes the CATCH opcode
func RunCatch(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	sym, sig, mode, b, err := ParseCatch(b)
	if err != nil {
		return b, err
	}
	r, err := matchFlag(st, sig, mode)
	if err != nil {
		return b, err
	}
	if r {
		log.Printf("catch at flag %v, moving to %v", sig, sym) //bitField, d)
		st.Down(sym)
		b = []byte{}
	} 
	return b, nil
}

// RunMap executes the CROAK opcode
func RunCroak(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	sig, mode, b, err := ParseCroak(b)
	if err != nil {
		return b, err
	}
	r, err := matchFlag(st, sig, mode)
	if err != nil {
		return b, err
	}
	if r {
		log.Printf("croak at flag %v, purging and moving to top", sig)
		st.Reset()
		b = []byte{}
	}
	return []byte{}, nil
}

// RunLoad executes the LOAD opcode
func RunLoad(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	sym, sz, b, err := ParseLoad(b)
	if err != nil {
		return b, err
	}

	r, err := refresh(sym, rs, ctx)
	if err != nil {
		return b, err
	}
	err = st.Add(sym, r, uint16(sz))
	return b, err
}

// RunLoad executes the RELOAD opcode
func RunReload(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	sym, b, err := ParseReload(b)
	if err != nil {
		return b, err
	}

	r, err := refresh(sym, rs, ctx)
	if err != nil {
		return b, err
	}
	st.Update(sym, r)
	return b, nil
}

// RunLoad executes the MOVE opcode
func RunMove(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	sym, b, err := ParseMove(b)
	if err != nil {
		return b, err
	}
	st.Down(sym)
	code, err := rs.GetCode(sym)
	if err != nil {
		return b, err
	}
	log.Printf("loaded additional code: %x", code)
	b = append(b, code...)
	return b, nil
}

// RunIncmp executes the INCMP opcode
func RunInCmp(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	sym, target, b, err := ParseInCmp(b)
	if err != nil {
		return b, err
	}
	v, err := st.GetFlag(state.FLAG_INMATCH)
	if err != nil {
		return b, err
	}
	if v {
		return b, nil
	}
	input, err := st.GetInput()
	if err != nil {
		return b, err
	}
	if sym == string(input) {
		log.Printf("input match for '%s'", input)
		_, err = st.SetFlag(state.FLAG_INMATCH)
		st.Down(target)
	}
	code, err := rs.GetCode(target)
	if err != nil {
		return b, err
	}
	log.Printf("loaded additional code: %x", code)
	b = append(b, code...)
	return b, err
}

// RunHalt executes the HALT opcode
func RunHalt(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	var err error
	b, err = ParseHalt(b)
	if err != nil {
		return b, err
	}
	log.Printf("found HALT, stopping")
	_, err = st.ResetFlag(state.FLAG_INMATCH)
	return b, err
}

// RunMSize
func RunMSize(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	log.Printf("WARNING MSIZE not yet implemented")
	_, _, b, err := ParseMSize(b)
	return b, err
}

// RunMSize
func RunMNext(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	selector, display, b, err := ParseMNext(b)
	if err != nil {
		return b, err
	}
	err = rs.SetMenuBrowse(selector, display, false)
	return b, err
}

// RunMSize
func RunMPrev(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	selector, display, b, err := ParseMPrev(b)
	if err != nil {
		return b, err
	}
	err = rs.SetMenuBrowse(selector, display, false)
	return b, err
}

func RunMOut(b []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	choice, title, b, err := ParseMOut(b)
	if err != nil {
		return b, err
	}
	err = rs.PutMenu(choice, title)
	return b, err
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
	return fn(ctx)
}
