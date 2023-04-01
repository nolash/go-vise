package vm

import (
	"encoding/binary"
)
const VERSION = 0

const (
	BACK = 0
	CATCH = 1
	CROAK = 2
	LOAD = 3
	RELOAD = 4
	MAP = 5
	MOVE = 6
	HALT = 7
	_MAX = 7
)

func NewLine(instructionList []byte, instruction uint16, args []string, post []byte, szPost []uint8) []byte {
	b := []byte{0x00, 0x00}
	binary.BigEndian.PutUint16(b, instruction)
	for _, arg := range args {
		b = append(b, uint8(len(arg)))
		b = append(b, []byte(arg)...)
	}
	if post != nil {
		b = append(b, uint8(len(post)))
		b = append(b, post...)
	}
	if szPost != nil {
		b = append(b, szPost...)
	}
	return append(instructionList, b...)
}
