package vm

import (
	"encoding/binary"
	"context"
	"fmt"
	"log"

	"git.defalsify.org/festive/resource"
	"git.defalsify.org/festive/router"
	"git.defalsify.org/festive/state"
)

//type Runner func(instruction []byte, st state.State, rs resource.Resource, ctx context.Context) (state.State, []byte, error)

func argFromBytes(input []byte) (string, []byte, error) {
	if len(input) == 0 {
		return "", input, fmt.Errorf("zero length input")
	}
	sz := input[0]
	out := input[1:1+sz]
	return string(out), input[1+sz:], nil
}

// Apply applies input to router bytecode to resolve the node symbol to execute.
//
// The execution byte code is initialized with the appropriate MOVE
//
// If the router indicates an argument input, the optional argument is set on the state.
//
// TODO: the bytecode load is a separate step so Run should be run separately.
func Apply(input []byte, instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	var err error

	log.Printf("running input %v against instruction %v", input, instruction)
	arg, input, err := argFromBytes(input)
	if err != nil {
		return input, err
	}

	rt := router.FromBytes(instruction)
	sym := rt.Get(arg)
	if sym == "" {
		sym = rt.Default()
		st.PutArg(arg)
	}

	if sym == "" {
		instruction = NewLine([]byte{}, MOVE, []string{"_catch"}, nil, nil)
	} else {
		instruction, err = rs.GetCode(sym)
		if err != nil {
			return instruction, err
		}

		if sym == "_" {
			instruction = NewLine([]byte{}, BACK, nil, nil, nil)
		} else {
			new_instruction := NewLine([]byte{}, MOVE, []string{sym}, nil, nil)
			instruction = append(new_instruction, instruction...)
		}
	}

	instruction, err = Run(instruction, st, rs, ctx)
	if err != nil {
		return instruction, err
	}
	return instruction, nil
}

// Run extracts individual op codes and arguments and executes them.
//
// Each step may update the state.
//
// On error, the remaining instructions will be returned. State will not be rolled back.
func Run(instruction []byte, st *state.State, rs resource.Resource, ctx context.Context) ([]byte, error) {
	var err error
	for len(instruction) > 0 {
		log.Printf("instruction is now %v", instruction)
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
		case HALT:
			log.Printf("found HALT, stopping")
			return instruction[2:], err
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
	if st.GetIndex(bitField) {
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

// split instruction into symbol and arguments
func instructionSplit(b []byte) (string, []byte, error) {
	if len(b) == 0 {
		return "", nil, fmt.Errorf("argument is empty")
	}
	sz := uint8(b[0])
	if sz == 0 {
		return "", nil, fmt.Errorf("zero-length argument")
	}
	tailSz := uint8(len(b))
	if tailSz < sz {
		return "", nil, fmt.Errorf("corrupt instruction, len %v less than symbol length: %v", tailSz, sz)
	}
	r := string(b[1:1+sz])
	return r, b[1+sz:], nil
}
