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
func CheckInput(input []byte) error {
	if !inputRegex.Match(input) {
		return fmt.Errorf("Input '%s' does not match input format /%s/", input, inputRegexStr)
	}
	return nil
}

// control characters for relative navigation.
func checkControl(input []byte) error {
	if !ctrlRegex.Match(input) {
		return fmt.Errorf("Input '%s' does not match 'control' format /%s/", input, ctrlRegexStr)
	}
	return nil
}

// CheckSym validates the given byte string as a node symbol.
func CheckSym(input []byte) error {
	if !symRegex.Match(input) {
		return fmt.Errorf("Input '%s' does not match 'sym' format /%s/", input, symRegexStr)
	}
	return nil
}

// route parsed target symbol to navigation state change method,
func applyTarget(target []byte, st *state.State, ctx context.Context) (string, uint16, error) {
	var err error
	var valid bool
	sym, idx := st.Where()

	err = CheckInput(target)
	if err == nil {
		valid = true
	}

	if !valid {
		err = CheckSym(target)
		if err == nil {
			valid = true
		}
	}

	if !valid {
		err = checkControl(target)
		if err == nil {
			valid = true
		}
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
