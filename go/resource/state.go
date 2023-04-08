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

func(s *StateResource) WithState(st *state.State) *StateResource {
	if s.st != nil {
		panic("state already set")
	}
	s.st = st
	return s
}
