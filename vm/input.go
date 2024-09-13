package vm

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/state"
)

var (
	inputRegexStr = "^\\+?[a-zA-Z0-9].*$"
	inputRegex = regexp.MustCompile(inputRegexStr)
	ctrlRegexStr = "^[><_^.]$"
	ctrlRegex = regexp.MustCompile(ctrlRegexStr)
	symRegexStr = "^[a-zA-Z0-9][a-zA-Z0-9_]+$"
	symRegex = regexp.MustCompile(symRegexStr)
)

var (
	preInputRegexStr = make(map[int]*regexp.Regexp)
)

// InvalidInputError indicates client input that was unhandled by the bytecode (INCMP fallthrough)
type InvalidInputError struct {
	input string
}

// NewInvalidInputError creates a new InvalidInputError
func NewInvalidInputError(input string) error {
	return InvalidInputError{input}
}

// Error implements the Error interface.
func(e InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input: '%s'", e.input)
}

func RegisterInputValidator(k int, v string) error {
	var ok bool
	var err error

	_, ok = preInputRegexStr[k]
	if ok {
		return fmt.Errorf("input checker with key '%d' already registered", k)
	}
	preInputRegexStr[k], err = regexp.Compile(v)
	return err
}

// CheckInput validates the given byte string as client input.
func ValidInput(input []byte) (int, error) {
	if inputRegex.Match(input) {
		return -1, nil
	}
	for k, v := range preInputRegexStr {
		logg.Tracef("custom check input", "i", k, "regex", v)
		if v.Match(input) {
			logg.Debugf("match custom check input", "i", k, "regex", v, "input", input)
			return k, nil
		}
	}
	return -2, fmt.Errorf("Input '%s' does not match any input format (default: /%s/)", input, inputRegexStr)
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
	if bytes.Equal(input, []byte("_catch")) {
		return nil
	}
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
		if err != nil {
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
func applyTarget(target []byte, st *state.State, ca cache.Memory, ctx context.Context) (string, uint16, error) {
	var err error
	sym, idx := st.Where()

	ok := valid(target)
	if !ok {
		return sym, idx, fmt.Errorf("invalid input: %s", target)
	}

	switch string(target) {
	case "_":
		sym, err = st.Up()
		if err != nil {
			return sym, idx, err
		}
		err = ca.Pop()
		if err != nil {
			return sym, idx, err
		}

	case ">":
		idx, err = st.Next()
		if err != nil {
			return sym, idx, err
		}
	case "<":
		idx, err = st.Previous()
		if err != nil {
			return sym, idx, err
		}
	case "^":
		notTop := true
		for notTop {
			notTop, err := st.Top()
			if notTop {
				break
			}
			sym, err = st.Up()
			if err != nil {
				return sym, idx, err
			}
			err = ca.Pop()
			if err != nil {
				return sym, idx, err
			}
		}
	case ".":
		st.Same()
		location, idx := st.Where()
		return location, idx, nil
	default:
		sym = string(target)
		err := st.Down(sym)
		if err != nil {
			return sym, idx, err
		}
		err = ca.Push()
		if err != nil {
			return sym, idx, err
		}
		idx = 0
	}
	return sym, idx, nil
}
