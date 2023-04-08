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
	ctrlInputRegexStr = "^[<>_]$"
	ctrlInputRegex = regexp.MustCompile(inputRegexStr)
)


func CheckInput(input []byte) error {
	if !inputRegex.Match(input) {
		return fmt.Errorf("Input '%s' does not match format /%s/", input, inputRegexStr)
	}
	return nil
}

func applyControlInput(input []byte, st *state.State, ctx context.Context) (string, error) {
	var err error
	sym, idx := st.Where()
	switch input[0] {
	case '_':
		sym, err = st.Up()
		if err != nil {
			return sym, err
		}
	}
	_ = idx
	return sym, nil
}

func ApplyInput(inputString string, st *state.State, ctx context.Context) (string, error) {
	input := []byte(inputString)
	if ctrlInputRegex.Match(input) {
		return applyControlInput(input, st, ctx)
	}

	err := CheckInput(input)
	if err != nil {
		return "", err
	}
	st.Down(inputString)
	return inputString, nil
}
