package vm

import (
	"fmt"
)


func ToString(b []byte) (string, error) {
	var s string
	running := true
	for running {
		op, bb, err := opSplit(b)
		b = bb
		if err != nil {
			return "", err
		}
		opString := OpcodeString[op]
		if opString == "" {
			return "", fmt.Errorf("unknown opcode: %v", op)
		}
		s += opString

		switch op {
		case CATCH:
			r, n, m, bb, err := ParseCatch(b)
			b = bb
			if err != nil {
				return "", err
			}
			vv := 0
			if m {
				vv = 1
			}
			s = fmt.Sprintf("%s %s %v %v # invertmatch=%v", s, r, n, vv, m)
		case CROAK:
			n, m, bb, err := ParseCroak(b)
			b = bb
			if err != nil {
				return "", err
			}
			vv := 0
			if m {
				vv = 1
			}
			s = fmt.Sprintf("%s %v %v # invertmatch=%v", s, n, vv, m)
		case LOAD:
			r, n, bb, err := ParseLoad(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s %v", s, r, n)
		case RELOAD:
			r, bb, err := ParseReload(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s", s, r)
		case MAP:
			r, bb, err := ParseMap(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s", s, r)
		case MOVE:
			r, bb, err := ParseMove(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s", s, r)
		case INCMP:
			r, v, bb, err := ParseInCmp(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s %s", s, r, v)
		case HALT:
			b, err = ParseHalt(b)
			if err != nil {
				return "", err
			}
		case MSIZE:
			r, v, bb, err := ParseMSize(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %v %v", s, r, v)
		case MOUT:
			r, v, bb, err := ParseMOut(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s \"%s\"", s, r, v)
		case MNEXT:
			r, v, bb, err := ParseMNext(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s \"%s\"", s, r, v)
		case MPREV:
			r, v, bb, err := ParseMPrev(b)
			b = bb
			if err != nil {
				return "", err
			}
			s = fmt.Sprintf("%s %s \"%s\"", s, r, v)
		}
		s += "\n"
		if len(b) == 0 {
			running = false
		}
	}
	return s, nil
}
