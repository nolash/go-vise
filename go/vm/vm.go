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
	default:
		err := fmt.Errorf("Unhandled state: %v", op)
		return st, err
	}
	return st, nil
}

func RunCatch(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}


func RunCroak(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}

func RunLoad(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}

func RunReload(instruction []byte, st state.State, rs resource.Fetcher, ctx context.Context) (state.State, error) {
	return st, nil
}
