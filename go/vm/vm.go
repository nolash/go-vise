package vm

import (
	"encoding/binary"
	"fmt"
	"context"

	"git.defalsify.org/festive/state"
	"git.defalsify.org/festive/resource"
)

type Runner func(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error)

func Run(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	op := binary.BigEndian.Uint16(instruction[:2])
	if op > _MAX {
		return st, fmt.Errorf("opcode value %v out of range (%v)", op, _MAX)
	}
	switch op {
	case CATCH:
		RunCatch(instruction[2:], st, rs, ctx)
	case CROAK:
		RunCroak(instruction[2:], st, rs, ctx)
	case LOAD:
		RunLoad(instruction[2:], st, rs, ctx)
	case RELOAD:
		RunReload(instruction[2:], st, rs, ctx)
	case MAP:
		RunMap(instruction[2:], st, rs, ctx)
	case SINK:
		RunSink(instruction[2:], st, rs, ctx)
	default:
		err := fmt.Errorf("Unhandled state: %v", op)
		return st, err
	}
	return st, nil
}

func instructionSplit(b []byte) (string, []byte, error) {
	sz := uint8(b[0])
	tailSz := uint8(len(b))
	if tailSz - 1 < sz {
		return "", nil, fmt.Errorf("corrupt instruction, len %v less than symbol length: %v", tailSz, sz)
	}
	r := string(b[1:1+sz])
	return r, b[1+sz:], nil
}

func RunMap(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, err
	}
	_ = tail
	st.Map(head)
	return st, nil
}

func RunSink(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}

func RunCatch(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}

func RunCroak(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}

func RunLoad(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	head, tail, err := instructionSplit(instruction)
	if err != nil {
		return st, err
	}
	fn, err := rs.FuncFor(head)
	if err != nil {
		return st, err
	}
	r, err := fn(tail, ctx)
	if err != nil {
		return st, err
	}
	st.Add(head, r)
	return st, nil
}

func RunReload(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}
