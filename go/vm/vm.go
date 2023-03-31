package vm

import (
	"encoding/binary"
	"fmt"
	"context"
	"log"

	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/resource"
)

type Runner func(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error)

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
		case SINK:
			st, instruction, err = RunSink(instruction[2:], st, rs, ctx)
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
	st.Map(head)
	return st, tail, nil
}

func RunSink(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	return st, nil, nil
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
	st.Add(head, r, uint32(len(r)))
	return st, tail, nil
}

func RunCroak(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	_ = head
	st.Reset()
	return st, tail, nil
}

func RunLoad(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, instruction, err
	}
	if !st.Check(head) {
		return st, instruction, fmt.Errorf("key %v already loaded", head)
	}
	sz := uint32(tail[0])
	tail = tail[1:]

	r, err := refresh(head, tail, rs, ctx)
	if err != nil {
		return st, tail, err
	}
	st.Add(head, r, sz)
	return st, tail, nil
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
	st.Add(head, r, uint32(len(r)))
	return st, tail, nil
}

func RunMove(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, []byte, error) {
	return st, nil, nil
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
