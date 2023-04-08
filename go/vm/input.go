package vm

import (
	"context"
	"fmt"
	"regexp"

	"git.defalsify.org/festive/state"
)

var (
	inputRegexStr = "^[a-zA-Z0-9].*$"
	inputRegex = regexp.MustCompile(inputRegexStr)
	ctrlRegexStr = "^[<>_]$"
	ctrlRegex = regexp.MustCompile(inputRegexStr)
	symRegexStr = "^[a-zA-Z0-9][a-zA-Z0-9_]+$"
	symRegex = regexp.MustCompile(inputRegexStr)

)

// CheckInput validates the given byte string as client input.
func ValidInput(input []byte) error {
	if !inputRegex.Match(input) {
		return fmt.Errorf("Input '%s' does not match input format /%s/", input, inputRegexStr)
	}
	return nil
}

// control characters for relative navigation.
func validControl(input []byte) error {
	if !ctrlRegex.Match(input) {
		return fmt.Errorf("Input '%s' does not match 'control' format /%s/", input, ctrlRegexStr)
	}
	return nil
}

// CheckSym validates the given byte string as a node symbol.
func ValidSym(input []byte) error {
	if !symRegex.Match(input) {
		return fmt.Errorf("Input '%s' does not match 'sym' format /%s/", input, symRegexStr)
	}
	return nil
}

// false if target is not valid
func valid(target []byte) bool {
	var ok bool
	if len(target) == 0 {
		return false
	}

	err := ValidSym(target)
	if err == nil {
		ok = true
	}

	if !ok {
		err = validControl(target)
		if err == nil {
			ok = true
		}
	}
	return ok 
}

// CheckTarget tests whether the navigation state transition is available in the current state.
//
// Fails if target is formally invalid, or if navigation is unavailable.
func CheckTarget(target []byte, st *state.State) (bool, error) {
	ok := valid(target)
	if !ok {
		return false, fmt.Errorf("invalid target: %x", target)
	}

	switch target[0] {
	case '_':
		topOk, err := st.Top()
		if err!= nil {
			return false, err
		}
		return topOk, nil
	case '<':
		_, prevOk := st.Sides()
		return prevOk, nil
	case '>':
		nextOk, _ := st.Sides()
		return nextOk, nil
	}
	return true, nil
}

// route parsed target symbol to navigation state change method,
func applyTarget(target []byte, st *state.State, ctx context.Context) (string, uint16, error) {
	var err error
	sym, idx := st.Where()

	ok := valid(target)
	if !ok {
		return sym, idx, fmt.Errorf("invalid input: %x", target)
	}

	switch target[0] {
	case '_':
		sym, err = st.Up()
		if err != nil {
			return sym, idx, err
		}
	case '>':
		idx, err = st.Next()
		if err != nil {
			return sym, idx, err
		}
	case '<':
		idx, err = st.Previous()
		if err != nil {
			return sym, idx, err
		}
	default:
		sym = string(target)
		st.Down(sym)
		idx = 0
	}
	return sym, idx, nil
}
