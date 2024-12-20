package vm

import (
	"bytes"
	"fmt"
	"io"
)

type ParseHandler struct {
	Catch func(string, uint32, bool) error
	Croak func(uint32, bool) error
	Load func(string, uint32) error
	Reload func(string) error
	Map func(string) error
	Move func(string) error
	Halt func() error
	InCmp func(string, string) error
	MOut func(string, string) error
	MSink func() error
	MNext func(string, string) error
	MPrev func(string, string) error
	cur string
	n int
	w io.Writer
}

func NewParseHandler() *ParseHandler {
	return &ParseHandler{}
}

func (ph *ParseHandler) Length() int {
	return ph.n
}

func (ph *ParseHandler) WithDefaultHandlers() *ParseHandler {
	ph.Catch = ph.catch
	ph.Croak = ph.croak
	ph.Load = ph.load
	ph.Reload = ph.reload
	ph.Map = ph.maph
	ph.Move = ph.move
	ph.Halt = ph.halt
	ph.InCmp = ph.incmp
	ph.MOut = ph.mout
	ph.MSink = ph.msink
	ph.MNext = ph.mnext
	ph.MPrev = ph.mprev
	return ph
}

func (ph *ParseHandler) WithWriter(w io.Writer) *ParseHandler {
	ph.w = w
	return ph
}

// TODO: output op sym
func (ph *ParseHandler) flush() error {
	if ph.w != nil {
		n, err := io.WriteString(ph.w, ph.cur)
		if err != nil {
			return err
		}
		ph.n += n
		ph.cur = ""
		//logg.Tracef("instruction debug write", "bytes", n, "instruction", s)
		logg.Tracef("instruction debug write", "bytes", n)
	}
	return nil
}

func (ph *ParseHandler) catch(sym string, flag uint32, inv bool) error {
	s := OpcodeString[CATCH]
	vv := 0
	if inv {
		vv = 1
	}
	ph.cur = fmt.Sprintf("%s %s %v %v\n", s, sym, flag, vv)
	return nil
}

func (ph *ParseHandler) croak(flag uint32, inv bool) error {
	s := OpcodeString[CROAK]
	vv := 0
	if inv {
		vv = 1
	}
	ph.cur = fmt.Sprintf("%s %v %v\n", s, flag, vv)
	return nil
}

func (ph *ParseHandler) load(sym string, length uint32) error {
	s := OpcodeString[LOAD]
	ph.cur = fmt.Sprintf("%s %s %v\n", s, sym, length)
	return nil
}

func (ph *ParseHandler) reload(sym string) error {
	s := OpcodeString[RELOAD]
	ph.cur = fmt.Sprintf("%s %s\n", s, sym)
	return nil
}

func (ph *ParseHandler) maph(sym string) error {
	s := OpcodeString[MAP]
	ph.cur = fmt.Sprintf("%s %s\n", s, sym)
	return nil
}

func (ph *ParseHandler) move(sym string) error {
	s := OpcodeString[MOVE]
	ph.cur = fmt.Sprintf("%s %s\n", s, sym)
	return nil
}

func (ph *ParseHandler) incmp(sym string, sel string) error {
	s := OpcodeString[INCMP]
	ph.cur = fmt.Sprintf("%s %s %v\n", s, sym, sel)
	return nil
}

func (ph *ParseHandler) halt() error {
	s := OpcodeString[HALT]
	ph.cur = fmt.Sprintf("%s\n", s)
	return nil
}

func (ph *ParseHandler) msink() error {
	s := OpcodeString[MSINK]
	ph.cur = fmt.Sprintf("%s\n", s)
	return nil
}

func (ph *ParseHandler) mout(sym string, sel string) error {
	s := OpcodeString[MOUT]
	ph.cur = fmt.Sprintf("%s %s %v\n", s, sym, sel)
	return nil
}

func (ph *ParseHandler) mnext(sym string, sel string) error {
	s := OpcodeString[MNEXT]
	ph.cur = fmt.Sprintf("%s %s %s\n", s, sym, sel)
	return nil
}

func (ph *ParseHandler) mprev(sym string, sel string) error {
	s := OpcodeString[MPREV]
	ph.cur = fmt.Sprintf("%s %s %s\n", s, sym, sel)
	return nil
}

// ToString verifies all instructions in bytecode and returns an assmebly code instruction for it.
func (ph *ParseHandler) ToString(b []byte) (string, error) {
	buf := bytes.NewBuffer(nil)
	ph = ph.WithWriter(buf)
	n, err := ph.ParseAll(b)
	if err != nil {
		return "", err
	}
	logg.Tracef("", "bytes_written", n)
	return buf.String(), nil
}

// ParseAll parses and verifies all instructions from bytecode.
//
// If writer is not nil, the parsed instruction as assembly code line string is written to it.
//
// Bytecode is consumed (and written) one instruction at a time.
//
// It fails on any parse error encountered before the bytecode EOF is reached.
func (ph *ParseHandler) ParseAll(b []byte) (int, error) {
	var s string
	running := true
	for running {
		op, bb, err := opSplit(b)
		b = bb
		if err != nil {
			return ph.Length(), err
		}
		s = OpcodeString[op]
		if s == "" {
			return ph.Length(), fmt.Errorf("unknown opcode: %v", op)
		}

		switch op {
		case CATCH:
			r, n, m, bb, err := ParseCatch(b)
			b = bb
			if err == nil {
				err = ph.Catch(r, n, m)
			}
		case CROAK:
			n, m, bb, err := ParseCroak(b)
			b = bb
			if err == nil {
				err = ph.Croak(n, m)
			}
		case LOAD:
			r, n, bb, err := ParseLoad(b)
			b = bb
			if err == nil {
				err = ph.Load(r, n)
			}
		case RELOAD:
			r, bb, err := ParseReload(b)
			b = bb
			if err == nil {
				err = ph.Reload(r)
			}
		case MAP:
			r, bb, err := ParseMap(b)
			b = bb
			if err == nil {
				err = ph.Map(r)
			}
		case MOVE:
			r, bb, err := ParseMove(b)
			b = bb
			if err == nil {
				err = ph.Move(r)
			}
		case INCMP:
			r, v, bb, err := ParseInCmp(b)
			b = bb
			if err == nil {
				err = ph.InCmp(r, v)
			}
		case HALT:
			b, err = ParseHalt(b)
			if err == nil {
				err = ph.Halt()
			}
		case MSINK:
			b, err = ParseMSink(b)
			if err == nil {
				err = ph.MSink()
			}
		case MOUT:
			r, v, bb, err := ParseMOut(b)
			b = bb
			if err == nil {
				err = ph.MOut(r, v)
			}
		case MNEXT:
			r, v, bb, err := ParseMNext(b)
			b = bb
			if err == nil {
				err = ph.MNext(r, v)
			}
		case MPREV:
			r, v, bb, err := ParseMPrev(b)
			b = bb
			if err == nil {
				err = ph.MPrev(r, v)
			}
		}
		if err != nil {
			return ph.Length(), err
		}
		ph.flush()

		//rs += "\n"
		if len(b) == 0 {
			running = false
		}
	}
	return ph.Length(), nil
}
