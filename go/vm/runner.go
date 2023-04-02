package vm

import (
	"encoding/binary"
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
func Run(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	var err error
	for len(instruction) > 0 {
		log.Printf("instruction is now 0x%x", instruction)
		op := binary.BigEndian.Uint16(instruction[:2])
		if op > _MAX {
			return instruction, fmt.Errorf("opcode value %v out of range (%v)", op, _MAX)
		}
		switch op {
		case CATCH:
			instruction, err = RunCatch(instruction[2:], st, rs, ctx)
		case CROAK:
			instruction, err = RunCroak(instruction[2:], st, rs, ctx)
		case LOAD:
			instruction, err = RunLoad(instruction[2:], st, rs, ctx)
		case RELOAD:
			instruction, err = RunReload(instruction[2:], st, rs, ctx)
		case MAP:
			instruction, err = RunMap(instruction[2:], st, rs, ctx)
		case MOVE:
			instruction, err = RunMove(instruction[2:], st, rs, ctx)
		case BACK:
			instruction, err = RunBack(instruction[2:], st, rs, ctx)
		case INCMP:
			instruction, err = RunIncmp(instruction[2:], st, rs, ctx)
		case HALT:
			return RunHalt(instruction[2:], st, rs, ctx)
		default:
			err = fmt.Errorf("Unhandled state: %v", op)
		}
		if err != nil {
			return instruction, err
		}
	}
	return instruction, nil
}

// RunMap executes the MAP opcode
func RunMap(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return instruction, err
	}
	err = st.Map(head)
	return tail, err
}

// RunMap executes the CATCH opcode
func RunCatch(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return instruction, err
	}
	bitFieldSize := tail[0]
	bitField := tail[1:1+bitFieldSize]
	tail = tail[1+bitFieldSize:]
	matchMode := tail[0] // matchmode 1 is match NOT set bit
	tail = tail[1:]
	match := false
	if matchMode > 0 {
		if !st.GetIndex(bitField) {
			match = true
		}
	} else if st.GetIndex(bitField) {
		match = true	
	}

	if match {
		log.Printf("catch at flag %v, moving to %v", bitField, head)
		st.Down(head)
		tail = []byte{}
	} 
	return tail, nil
}

// RunMap executes the CROAK opcode
func RunCroak(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return instruction, err
	}
	_ = head
	_ = tail
	st.Reset()
	return []byte{}, nil
}

// RunLoad executes the LOAD opcode
func RunLoad(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return instruction, err
	}
	if !st.Check(head) {
		return instruction, fmt.Errorf("key %v already loaded", head)
	}
	sz := uint16(tail[0])
	tail = tail[1:]

	r, err := refresh(head, rs, ctx)
	if err != nil {
		return tail, err
	}
	err = st.Add(head, r, sz)
	return tail, err
}

// RunLoad executes the RELOAD opcode
func RunReload(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return instruction, err
	}
	r, err := refresh(head, rs, ctx)
	if err != nil {
		return tail, err
	}
	st.Update(head, r)
	return tail, nil
}

// RunLoad executes the MOVE opcode
func RunMove(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return instruction, err
	}
	st.Down(head)
	return tail, nil
}

// RunLoad executes the BACK opcode
func RunBack(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	st.Up()
	return instruction, nil
}

// RunIncmp executes the INCMP opcode
func RunIncmp(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return instruction, err
	}
	sym, tail, err := instructionSplit(tail)
	if err != nil {
		return instruction, err
	}
	v, err := st.GetFlag(state.FLAG_INMATCH)
	if err != nil {
		return tail, err
	}
	if v {
		return tail, nil
	}
	input, err := st.GetInput()
	if err != nil {
		return tail, err
	}
	log.Printf("checking input %v %v", input, head)
	if head == string(input) {
		log.Printf("input match for '%s'", input)
		_, err = st.SetFlag(state.FLAG_INMATCH)
		st.Down(sym)
	}
	return tail, err
}

// RunHalt executes the HALT opcode
func RunHalt(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	log.Printf("found HALT, stopping")
	_, err := st.ResetFlag(state.FLAG_INMATCH)
	return instruction, err
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
