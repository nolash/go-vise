package vm

const VERSION = 0

type Opcode uint16

// VM Opcodes
const (
	NOOP = 0
	CATCH = 1
	CROAK = 2
	LOAD = 3
	RELOAD = 4
	MAP = 5
	MOVE = 6
	HALT = 7
	INCMP = 8
	MSINK = 9
	MOUT = 10
	MNEXT = 11
	MPREV = 12
	_MAX = 12
)

var (
	OpcodeString = map[Opcode]string{
		NOOP: "NOOP",
		CATCH: "CATCH",
		CROAK: "CROAK",
		LOAD: "LOAD",
		RELOAD: "RELOAD",
		MAP: "MAP",
		MOVE: "MOVE",
		HALT: "HALT",
		INCMP: "INCMP",
		MSINK: "MSINK",
		MOUT: "MOUT",
		MNEXT: "MNEXT",
		MPREV: "MPREV",
	}

	OpcodeIndex = map[string]Opcode {
		"NOOP": NOOP,
		"CATCH": CATCH,
		"CROAK": CROAK,
		"LOAD": LOAD,
		"RELOAD": RELOAD,
		"MAP": MAP,
		"MOVE": MOVE,
		"HALT": HALT,
		"INCMP": INCMP,
		"MSINK": MSINK,
		"MOUT": MOUT,
		"MNEXT": MNEXT,
		"MPREV": MPREV,
	}

)
