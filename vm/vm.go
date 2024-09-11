package vm

import (
	"encoding/binary"
	"fmt"
)

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

// ParseOp verifies and extracts the expected opcode portion of an instruction
func ParseOp(b []byte) (Opcode, []byte, error) {
	op, b, err := opSplit(b)
	if err != nil {
		return NOOP, b, err
	}
	return op, b, nil
}

// ParseLoad parses and extracts the expected argument portion of a LOAD instruction
func ParseLoad(b []byte) (string, uint32, []byte, error) {
	return parseSymLen(b)
}

// ParseReload parses and extracts the expected argument portion of a RELOAD instruction
func ParseReload(b []byte) (string, []byte, error) {
	return parseSym(b)
}

// ParseMap parses and extracts the expected argument portion of a MAP instruction
func ParseMap(b []byte) (string, []byte, error) {
	return parseSym(b)
}

// ParseMove parses and extracts the expected argument portion of a MOVE instruction
func ParseMove(b []byte) (string, []byte, error) {
	return parseSym(b)
}

// ParseHalt parses and extracts the expected argument portion of a HALT instruction
func ParseHalt(b []byte) ([]byte, error) {
	return parseNoArg(b)
}

// ParseCatch parses and extracts the expected argument portion of a CATCH instruction
func ParseCatch(b []byte) (string, uint32, bool, []byte, error) {
	return parseSymSig(b)
}

// ParseCroak parses and extracts the expected argument portion of a CROAK instruction
func ParseCroak(b []byte) (uint32, bool, []byte, error) {
	return parseSig(b)
}

// ParseInCmp parses and extracts the expected argument portion of a INCMP instruction
func ParseInCmp(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

// ParseMPrev parses and extracts the expected argument portion of a MPREV instruction
func ParseMPrev(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

// ParseMNext parses and extracts the expected argument portion of a MNEXT instruction
func ParseMNext(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

// ParseMSink parses and extracts the expected argument portion of a MSINK instruction
func ParseMSink(b []byte) ([]byte, error) {
	return parseNoArg(b)
//	if len(b) < 2 {
//		return 0, 0, b, fmt.Errorf("argument too short")
//	}
//	r := uint32(b[0])
//	rr := uint32(b[1])
//	b = b[2:]
//	return b, nil
}

// ParseMOut parses and extracts the expected argument portion of a MOUT instruction
func ParseMOut(b []byte) (string, string, []byte, error) {
	return parseTwoSym(b)
}

// noop
func parseNoArg(b []byte) ([]byte, error) {
	return b, nil
}

// parse and extract two length-prefixed string values
func parseSym(b []byte) (string, []byte, error) {
	sym, b, err := instructionSplit(b)
	if err != nil {
		return "", b, err
	}
	return sym, b, nil
}

// parse and extract two length-prefixed string values
func parseTwoSym(b []byte) (string, string, []byte, error) {
	symOne, b, err := instructionSplit(b)
	if err != nil {
		return "", "", b, err
	}
	symTwo, b, err := instructionSplit(b)
	if err != nil {
		return "", "", b, err
	}
	return symOne, symTwo, b, nil
}

// parse and extract one length-prefixed string value, and one length-prefixed integer value
func parseSymLen(b []byte) (string, uint32, []byte, error) {
	sym, b, err := instructionSplit(b)
	if err != nil {
		return "", 0, b, err
	}
	sz, b, err := intSplit(b)
	if err != nil {
		return "", 0, b, err
	}
	return sym, sz, b, nil
}

// parse and extract one length-prefixed string value, and one single byte of integer
func parseSymSig(b []byte) (string, uint32, bool, []byte, error) {
	sym, b, err := instructionSplit(b)
	if err != nil {
		return "", 0, false, b, err
	}
	sig, b, err := intSplit(b)
	if err != nil {
		return "", 0, false, b, err
	}
	if len(b) == 0 {
		return "", 0, false, b, fmt.Errorf("instruction too short")
	}
	matchmode := b[0] > 0
	b = b[1:]
	
	return sym, sig, matchmode, b, nil
}

// parse and extract one single byte of integer
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

// split bytecode into head and b using length-prefixed integer
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

// split bytecode into head and b using length-prefixed string
func instructionSplit(b []byte) (string, []byte, error) {
	if len(b) == 0 {
		return "", nil, fmt.Errorf("argument is empty")
	}
	sz := uint8(b[0])
	if sz == 0 {
		return "", nil, fmt.Errorf("zero-length argument")
	}
	bSz := len(b)
	if bSz < int(sz) {
		return "", nil, fmt.Errorf("corrupt instruction, len %v less than symbol length: %v", bSz, sz)
	}
	r := string(b[1:1+sz])
	return r, b[1+sz:], nil
}

// split bytecode into head and b using opcode
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
