package resource

import (
	"git.defalsify.org/festive/state"
)

type StateResource struct {
	Resource
	st *state.State
}

func ToStateResource(rs Resource) *StateResource {
	return &StateResource{rs, nil}
}

func NewStateResource(st *state.State) *StateResource {
	return &StateResource {
		NewMenuResource(),
		st,
	}
}

func(sr *StateResource) WithState(st *state.State) *StateResource {
	if sr.st != nil {
		panic("state already set")
	}
	sr.st = st
	return sr
}