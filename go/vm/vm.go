package vm

import (
	"encoding/binary"
	"fmt"

	"git.defalsify.org/festive/state"
)

func ParseOp(b []byte) (Opcode, []byte, error) {
	op, b, err := opSplit(b)
	if err != nil {
		return NOOP, b, err
	}
	return op, b, nil
}

func ParseLoad(b []byte) (string, uint32, []byte, error) {
	return parseSymLen(b)
}

func ParseReload(b []byte) (string, []byte, error) {
	return parseSym(b)
}

func ParseMap(b []byte) (string, []byte, error) {
	return parseSym(b)
}

func ParseMove(b []byte) (string, []byte, error) {
	return parseSym(b)
}

func ParseHalt(b []byte) ([]byte, error) {
	return parseNoArg(b)
}

func ParseCatch(b []byte) (string, uint32, bool, []byte, error) {
	return parseSymSig(b)
}

func ParseCroak(b []byte) (uint32, bool, []byte, error) {
	return parseSig(b)
}

func ParseInCmp(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

func ParseMPrev(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

func ParseMNext(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

func ParseMSize(b []byte) (uint32, []byte, error) {
	if len(b) < 1 {
		return 0, b, fmt.Errorf("zero-length argument")
	}
	r := uint32(b[0])
	b = b[1:]
	return r, b, nil
}

func ParseMOut(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

func parseNoArg(b []byte) ([]byte, error) {
	return b, nil
}

func parseSym(b []byte) (string, []byte, error) {
	sym, tail, err := instructionSplit(b)
	if err != nil {
		return "", b, err
	}
	return sym, tail, nil
}

func parseTwoSym(b []byte) (string, string, []byte, error) {
	symOne, tail, err := instructionSplit(b)
	if err != nil {
		return "", "", b, err
	}
	symTwo, tail, err := instructionSplit(tail)
	if err != nil {
		return "", "", tail, err
	}
	return symOne, symTwo, tail, nil
}

func parseSymLen(b []byte) (string, uint32, []byte, error) {
	sym, tail, err := instructionSplit(b)
	if err != nil {
		return "", 0, b, err
	}
	sz, tail, err := intSplit(tail)
	if err != nil {
		return "", 0, b, err
	}
	return sym, sz, tail, nil
}

func parseSymSig(b []byte) (string, uint32, bool, []byte, error) {
	sym, tail, err := instructionSplit(b)
	if err != nil {
		return "", 0, false, b, err
	}
	sig, tail, err := intSplit(tail)
	if err != nil {
		return "", 0, false, b, err
	}
	if len(tail) == 0 {
		return "", 0, false, b, fmt.Errorf("instruction too short")
	}
	matchmode := tail[0] > 0
	tail = tail[1:]
	
	return sym, sig, matchmode, tail, nil
}

func parseSig(b []byte) (uint32, bool, []byte, error) {
	sig, b, err := intSplit(b)
	if err != nil {
		return 0, false, b, err
	}
	if len(b) == 0 {
		return 0, false, b, fmt.Errorf("instruction too short")
	}
	matchmode := b[0] > 0
	b = b[1:]
	
	return sig, matchmode, b, nil
}

func matchFlag(st *state.State, sig uint32, invertMatch bool) (bool, error) {
	r, err := st.GetFlag(sig)
	if err != nil {
		return false, err
	}
	if invertMatch {
		if !r {
			return true, nil
		}
	} else if r {
		return true, nil
	}
	return false, nil
}

// NewLine creates a new instruction line for the VM.
func NewLine(instructionList []byte, instruction uint16, strargs []string, byteargs []byte, numargs []uint8) []byte {
	if instructionList == nil {
		instructionList = []byte{}
	}
	b := []byte{0x00, 0x00}
	binary.BigEndian.PutUint16(b, instruction)
	for _, arg := range strargs {
		b = append(b, uint8(len(arg)))
		b = append(b, []byte(arg)...)
	}
	if byteargs != nil {
		b = append(b, uint8(len(byteargs)))
		b = append(b, byteargs...)
	}
	if numargs != nil {
		b = append(b, numargs...)
	}
	return append(instructionList, b...)
}

func byteSplit(b []byte) ([]byte, []byte, error) {
	bitFieldSize := b[0]
	bitField := b[1:1+bitFieldSize]
	b = b[1+bitFieldSize:]
	return bitField, b, nil
}

func intSplit(b []byte) (uint32, []byte, error) {
	l := uint8(b[0])
	sz := uint32(l)
	b = b[1:]
	if l > 0 {
		r := []byte{0, 0, 0, 0}
		c := 0
		ll := 4 - l
		var i uint8
		for i = 0; i < 4; i++ {
			if i >= ll {
				r[i] = b[c]
				c += 1
			}
		}
		sz = binary.BigEndian.Uint32(r)
		b = b[l:]
	}
	return sz, b, nil
}

// split instruction into symbol and arguments
func instructionSplit(b []byte) (string, []byte, error) {
	if len(b) == 0 {
		return "", nil, fmt.Errorf("argument is empty")
	}
	sz := uint8(b[0])
	if sz == 0 {
		return "", nil, fmt.Errorf("zero-length argument")
	}
	tailSz := uint8(len(b))
	if tailSz < sz {
		return "", nil, fmt.Errorf("corrupt instruction, len %v less than symbol length: %v", tailSz, sz)
	}
	r := string(b[1:1+sz])
	return r, b[1+sz:], nil
}

func opCheck(b []byte, opIn Opcode) ([]byte, error) {
	var bb []byte
	op, bb, err := opSplit(b)
	if err != nil {
		return b, err
	}
	b = bb
	if op != opIn {
		return b, fmt.Errorf("not a %v instruction", op)
	}
	return b, nil
}

func opSplit(b []byte) (Opcode, []byte, error) {
	l := len(b)
	if l < 2 {
		return 0, b, fmt.Errorf("input size %v too short for opcode", l)
	}
	op := binary.BigEndian.Uint16(b)
	if op > _MAX {
		return 0, b, fmt.Errorf("invalid opcode %v", op)
	}
	return Opcode(op), b[2:], nil
}
