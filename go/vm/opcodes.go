package vm

import (
	"encoding/binary"
)
const VERSION = 0

// Opcodes
const (
	BACK = 0
	CATCH = 1
	CROAK = 2
	LOAD = 3
	RELOAD = 4
	MAP = 5
	MOVE = 6
	HALT = 7
	INCMP = 8
	//IN = 9
	_MAX = 8
)

// NewLine creates a new instruction line for the VM.
func NewLine(instructionList []byte, instruction uint16, strargs []string, byteargs []byte, numargs []uint8) []byte {
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
