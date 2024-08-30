package asm

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"git.defalsify.org/vise.git/state"
)

// FlagParser is used to resolve flag strings to corresponding
// flag index integer values.
type FlagParser struct {
	flag map[string]string
	flagDescription map[uint32]string
	hi uint32
}

// NewFlagParser creates a new FlagParser
func NewFlagParser() *FlagParser {
	return &FlagParser{
		flag: make(map[string]string),
		flagDescription: make(map[uint32]string),
	}
}

// GetFlag returns the flag index value for a given flag string
// as a numeric string.
//
// If flag string has not been registered, an error is returned.
func(pp *FlagParser) GetAsString(key string) (string, error) {
	v, ok := pp.flag[key]
	if !ok {
		return "", fmt.Errorf("no flag registered under key: %s", key)
	}
	return v, nil
}

// GetFlag returns the flag index integer value for a given
// flag string
//
// If flag string has not been registered, an error is returned.
func(pp *FlagParser) GetFlag(key string) (uint32, error) {
	v, err := pp.GetAsString(key)
	if err != nil {
		return 0, err
	}
	r, err := strconv.Atoi(v) // cannot fail
	return uint32(r), nil
}

// GetDescription returns a flag description for a given flag index,
// if available.
//
// If no description has been provided, an error is returned.
func(pp *FlagParser) GetDescription(idx uint32) (string, error) {
	v, ok := pp.flagDescription[idx]
	if !ok {
		return "", fmt.Errorf("no description for flag idx: %v", idx)
	}
	return v, nil
}

// Last returns the highest registered flag index value
func(pp *FlagParser) Last() uint32 {
	return pp.hi	
}

// Load parses a Comma Seperated Value file under the given filepath
// to provide mappings between flag strings and flag indices.
//
// The expected format is:
// 
// Field 1: The literal string "flag"
// Field 2: Flag string
// Field 3: Flag index
// Field 4: Flag description (optional)
func(pp *FlagParser) Load(fp string) (int, error) {
	var i int
	f, err := os.Open(fp)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	for i = 0; true; i++ {
		v, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		if v[0] == "flag" {
			if len(v) < 3 {
				return 0, fmt.Errorf("Not enough fields for flag setting in line %d", i)
			}
			vv, err := strconv.Atoi(v[2])
			if err != nil {
				return 0, fmt.Errorf("Flag translation value must be numeric")
			}
			if vv < state.FLAG_USERSTART {
				return 0, fmt.Errorf("Minimum flag value is FLAG_USERSTART (%d)", FLAG_USERSTART)
			}
			fl := uint32(vv)
			pp.flag[v[1]] = v[2]
			if fl > pp.hi {
				pp.hi = fl
			}
			
			if (len(v) > 3) {
				pp.flagDescription[uint32(fl)] = v[3]
				Logg.Debugf("added flag translation", "from", v[1], "to", v[2], "description", v[3])
			} else {
				Logg.Debugf("added flag translation", "from", v[1], "to", v[2])
			}
		}
	}	

	return i, nil
}
