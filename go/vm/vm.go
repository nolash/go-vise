package vm

import (
	"encoding/binary"
	"fmt"
)


func ParseLoad(b []byte) (string, uint32, []byte, error) {
	return parseSymLen(b, LOAD)
}

func ParseReload(b []byte) (string, []byte, error) {
	return parseSym(b, RELOAD)
}

func ParseMap(b []byte) (string, []byte, error) {
	return parseSym(b, MAP)
}

func ParseMove(b []byte) (string, []byte, error) {
	return parseSym(b, MOVE)
}

func ParseHalt(b []byte) ([]byte, error) {
	return parseNoArg(b, HALT)
}

func ParseCatch(b []byte) (string, uint8, []byte, error) {
	return parseSymSig(b, CATCH)
}

func ParseCroak(b []byte) (string, uint8, []byte, error) {
	return parseSymSig(b, CROAK)
}

func ParseInCmp(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b, INCMP)
}

func parseNoArg(b []byte, op Opcode) ([]byte, error) {
	return opCheck(b, op)
}

func parseSym(b []byte, op Opcode) (string, []byte, error) {
	b, err := opCheck(b, op)
	if err != nil {
		return "", b, err
	}
	sym, tail, err := instructionSplit(b)
	if err != nil {
		return "", b, err
	}
	return sym, tail, nil
}

func parseTwoSym(b []byte, op Opcode) (string, string, []byte, error) {
	b, err := opCheck(b, op)
	if err != nil {
		return "", "", b, err
	}
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

func parseSymLen(b []byte, op Opcode) (string, uint32, []byte, error) {
	b, err := opCheck(b, op)
	if err != nil {
		return "", 0, b, err
	}
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

func parseSymSig(b []byte, op Opcode) (string, uint8, []byte, error) {
	b, err := opCheck(b, op)
	if err != nil {
		return "", 0, b, err
	}
	sym, tail, err := instructionSplit(b)
	if err != nil {
		return "", 0, b, err
	}
	if len(tail) == 0 {
		return "", 0, b, fmt.Errorf("instruction too short")
	}
	n := tail[0]
	tail = tail[1:]
	return sym, n, tail, nil
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
	op, b, err := opSplit(b)
	if err != nil {
		return b, err
	}
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
