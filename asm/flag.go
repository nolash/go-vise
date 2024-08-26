package asm

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type FlagParser struct {
	flag map[string]string
	flagDescription map[uint32]string
}

func NewFlagParser() *FlagParser {
	return &FlagParser{
		flag: make(map[string]string),
		flagDescription: make(map[uint32]string),
	}
}

func(pp *FlagParser) GetAsString(key string) (string, error) {
	v, ok := pp.flag[key]
	if !ok {
		return "", fmt.Errorf("no flag registered under key: %s", key)
	}
	return v, nil
}

func(pp *FlagParser) GetFlag(key string) (uint32, error) {
	v, err := pp.GetAsString(key)
	if err != nil {
		return 0, err
	}
	r, err := strconv.Atoi(v) // cannot fail
	return uint32(r), nil
}

func(pp *FlagParser) GetDescription(idx uint32) (string, error) {
	v, ok := pp.flagDescription[idx]
	if !ok {
		return "", fmt.Errorf("no description for flag idx: %v", idx)
	}
	return v, nil
}

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
			fl, err := strconv.Atoi(v[2])
			if err != nil {
				return 0, fmt.Errorf("Flag translation value must be numeric")
			}
			pp.flag[v[1]] = v[2]
			
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


