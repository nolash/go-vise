package vm

const VERSION = 0

type Opcode uint16

// VM Opcodes
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
	_MAX = 8
)
