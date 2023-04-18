package vm

import (
	"bytes"
	"fmt"
	"io"
)

// ToString verifies all instructions in bytecode and returns an assmebly code instruction for it.
func ToString(b []byte) (string, error) {
	buf := bytes.NewBuffer(nil)
	n, err := ParseAll(b, buf)
	if err != nil {
		return "", err
	}
	Logg.Tracef("", "bytes_written", n)
	return buf.String(), nil
}

// ParseAll parses and verifies all instructions from bytecode.
//
// If writer is not nil, the parsed instruction as assembly code line string is written to it.
//
// Bytecode is consumed (and written) one instruction at a time.
//
// It fails on any parse error encountered before the bytecode EOF is reached.
func ParseAll(b []byte, w io.Writer) (int, error) {
	var s string
	var rs string
	var rn int
	running := true
	for running {
		op, bb, err := opSplit(b)
		b = bb
		if err != nil {
			return rn, err
		}
		s = OpcodeString[op]
		if s == "" {
			return rn, fmt.Errorf("unknown opcode: %v", op)
		}

		switch op {
		case CATCH:
			r, n, m, bb, err := ParseCatch(b)
			b = bb
			if err == nil {
				if w != nil {
					vv := 0
					if m {
						vv = 1
					}
					if w != nil {
						//rs = fmt.Sprintf("%s %s %v %v # invertmatch=%v\n", s, r, n, vv, m)
						rs = fmt.Sprintf("%s %s %v %v\n", s, r, n, vv)
					}
				}
			}
		case CROAK:
			n, m, bb, err := ParseCroak(b)
			b = bb
			if err == nil {
				if w != nil {
					vv := 0
					if m {
						vv = 1
					}
					//rs = fmt.Sprintf("%s %v %v # invertmatch=%v\n", s, n, vv, m)
					rs = fmt.Sprintf("%s %v %v\n", s, n, vv)
				}
			}
		case LOAD:
			r, n, bb, err := ParseLoad(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s %v\n", s, r, n)
				}
			}
		case RELOAD:
			r, bb, err := ParseReload(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s\n", s, r)
				}
			}
		case MAP:
			r, bb, err := ParseMap(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s\n", s, r)
				}
			}
		case MOVE:
			r, bb, err := ParseMove(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s\n", s, r)
				}
			}
		case INCMP:
			r, v, bb, err := ParseInCmp(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s %s\n", s, r, v)
				}
			}
		case HALT:
			b, err = ParseHalt(b)
			rs = "HALT\n"
		case MSIZE:
			r, v, bb, err := ParseMSize(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %v %v\n", s, r, v)
				}
			}
		case MOUT:
			r, v, bb, err := ParseMOut(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s \"%s\"\n", s, r, v)
				}
			}
		case MNEXT:
			r, v, bb, err := ParseMNext(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s \"%s\"\n", s, r, v)
				}
			}
		case MPREV:
			r, v, bb, err := ParseMPrev(b)
			b = bb
			if err == nil {
				if w != nil {
					rs = fmt.Sprintf("%s %s \"%s\"\n", s, r, v)
				}
			}
		}
		if err != nil {
			return rn, err
		}
		if w != nil {
			n, err := io.WriteString(w, rs)
			if err != nil {
				return rn, err
			}
			rn += n
			Logg.Tracef("instruction debug write", "bytes", n, "instruction", s)
		}

		//rs += "\n"
		if len(b) == 0 {
			running = false
		}
	}
	return rn, nil
}
