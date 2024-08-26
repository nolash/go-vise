package asm

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type PreProcessor struct {
	flags map[string]string
}

func NewPreProcessor() *PreProcessor {
	return &PreProcessor{
		flags: make(map[string]string),
	}
}

func(pp *PreProcessor) Get(key string) (string, error) {
	v, ok := pp.flags[key]
	if !ok {
		return "", fmt.Errorf("no flag registered under key: %s", key)
	}
	return v, nil
}

func(pp *PreProcessor) Load(fp string) (int, error) {
	var i int
	f, err := os.Open(fp)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	for i = 0; true; i++ {
		r, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		if r[0] == "flag" {
			if len(r) < 3 {
				return 0, fmt.Errorf("Not enough fields for flag setting in line %d", i)
			}
			_, err = strconv.Atoi(r[2])
			if err != nil {
				return 0, fmt.Errorf("Flag translation value must be numeric")
			}
			pp.flags[r[1]] = r[2]
			Logg.Debugf("added flag translation", "from", r[1], "to", r[2])
		}
	}	

	return i, nil
}


