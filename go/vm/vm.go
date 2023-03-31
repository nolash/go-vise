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

//type Runner func(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error)

func argFromBytes(input []byte) (string, []byte, error) {
	if len(input) == 0 {
		return "", input, fmt.Errorf("zero length input")
	}
	sz := input[0]
	out := input[1:1+sz]
	return string(out), input[1+sz:], nil
}

func Apply(input []byte, instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	var err error

	arg, input, err := argFromBytes(input)
	if err != nil {
		return st, input, err
	}

	rt := router.FromBytes(input)
	sym := rt.Get(arg)
	if sym == "" {
		sym = rt.Default()
		st.PutArg(arg)
	} 
	if sym == "" {
		instruction = NewLine([]byte{}, MOVE, []string{"_catch"}, nil, nil)
	} else if sym == "_" {
		instruction = NewLine([]byte{}, BACK, nil, nil, nil)
	} else {
		new_instruction := NewLine([]byte{}, MOVE, []string{sym}, nil, nil)
		instruction = append(new_instruction, instruction...)
	}

	st, instruction, err = Run(instruction, st, rs, ctx)
	if err != nil {
		return st, instruction, err
	}
	return st, instruction, nil
}

func Run(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	var err error
	for len(instruction) > 0 {
		log.Printf("instruction is now %v", instruction)
		op := binary.BigEndian.Uint16(instruction[:2])
		if op > _MAX {
			return st, instruction, fmt.Errorf("opcode value %v out of range (%v)", op, _MAX)
		}
		switch op {
		case CATCH:
			st, instruction, err = RunCatch(instruction[2:], st, rs, ctx)
			break
		case CROAK:
			st, instruction, err = RunCroak(instruction[2:], st, rs, ctx)
			break
		case LOAD:
			st, instruction, err = RunLoad(instruction[2:], st, rs, ctx)
			break
		case RELOAD:
			st, instruction, err = RunReload(instruction[2:], st, rs, ctx)
			break
		case MAP:
			st, instruction, err = RunMap(instruction[2:], st, rs, ctx)
			break
		case MOVE:
			st, instruction, err = RunMove(instruction[2:], st, rs, ctx)
			break
		case BACK:
			st, instruction, err = RunBack(instruction[2:], st, rs, ctx)
			break
		default:
			err = fmt.Errorf("Unhandled state: %v", op)
		}
		if err != nil {
			return st, instruction, err
		}
	}
	return st, instruction, nil
}

func RunMap(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	err = st.Map(head)
	return st, tail, err
}

func RunCatch(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	r, err := rs.Get(head)
	if err != nil {
		return st, instruction, err
	}
	_ = tail
	st.Add(head, r, 0)
	return st, []byte{}, nil
}

func RunCroak(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	_ = head
	_ = tail
	st.Reset()
	return st, []byte{}, nil
}

func RunLoad(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	if !st.Check(head) {
		return st, instruction, fmt.Errorf("key %v already loaded", head)
	}
	sz := uint16(tail[0])
	tail = tail[1:]

	r, err := refresh(head, tail, rs, ctx)
	if err != nil {
		return st, tail, err
	}
	err = st.Add(head, r, sz)
	return st, tail, err
}

func RunReload(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	r, err := refresh(head, tail, rs, ctx)
	if err != nil {
		return st, tail, err
	}
	st.Update(head, r)
	return st, tail, nil
}

func RunMove(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	st.Down(head)
	return st, tail, nil
}

func RunBack(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	st.Up()
	return st, instruction, nil
}

func refresh(key string, sym []byte, rs resource.Fetcher, ctx context.Context) (string, error) {
	fn, err := rs.FuncFor(key)
	if err != nil {
		return "", err
	}
	return fn(sym, ctx)
}

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
