package state

import (
	"fmt"
	"strings"

	"git.defalsify.org/vise.git/lang"
)

type IndexError struct {
}

func(err *IndexError) Error() string {
	return fmt.Sprintf("already at first index")
}

// State holds the command stack, error condition of a unique execution session.
//
// It also holds cached values for all results of executed symbols.
//
// Cached values are linked to the command stack level it which they were loaded. When they go out of scope they are freed.
//
// Values must be mapped to a level in order to be available for retrieval and count towards size
//
// It can hold a single argument, which is freed once it is read
//
// Symbols are loaded with individual size limitations. The limitations apply if a load symbol is updated. Symbols may be added with a 0-value for limits, called a "sink." If mapped, the sink will consume all net remaining size allowance unused by other symbols. Only one sink may be mapped per level.
//
// Symbol keys do not count towards cache size limitations.
//
// 8 first flags are reserved.
type State struct {
	Code []byte // Pending bytecode to execute
	ExecPath []string // Command symbols stack
	BitSize uint32 // Size of (32-bit capacity) bit flag byte array
	SizeIdx uint16 // Lateral page browse index in current frame
	Flags []byte // Error state
	Moves uint32 // Number of times navigation has been performed
	Language *lang.Language // Language selector for rendering
	input []byte // Last input
	debug bool // Make string representation more human friendly
}

// number of bytes necessary to represent a bitfield of the given size.
func toByteSize(BitSize uint32) uint8 {
	if BitSize == 0 {
		return 0
	}
	n := BitSize % 8
	if n > 0 {
		BitSize += (8 - n)
	}
	return uint8(BitSize / 8)
}

//// Retrieve the state of a state flag
//func getFlag(bitIndex uint32, bitField []byte) bool {
//	byteIndex := bitIndex / 8
//	localBitIndex := bitIndex % 8
//	b := bitField[byteIndex]
//	return (b & (1 << localBitIndex)) > 0
//}

// NewState creates a new State object with BitSize number of error condition states in ADDITION to the 8 builtin flags.
func NewState(BitSize uint32) State {
	st := State{
		BitSize: BitSize + 8,
	}
	byteSize := toByteSize(BitSize + 8)
	if byteSize > 0 {
		st.Flags = make([]byte, byteSize) 
	} else {
		st.Flags = []byte{}
	}
	return st
}

func(st State) WithDebug() State {
	st.debug = true
	return st
}

// SetFlag sets the flag at the given bit field index
//
// Returns true if bit state was changed.
//
// Fails if bitindex is out of range.
func(st *State) SetFlag(bitIndex uint32) (bool, error) {
	if bitIndex + 1 > st.BitSize {
		return false, fmt.Errorf("bit index %v is out of range of bitfield size %v", bitIndex, st.BitSize)
	}
	r := getFlag(bitIndex, st.Flags)
	if r {
		return false, nil
	}
	byteIndex := bitIndex / 8
	localBitIndex := bitIndex % 8
	b := st.Flags[byteIndex] 
	st.Flags[byteIndex] = b | (1 << localBitIndex)
	return true, nil
}


// ResetFlag resets the flag at the given bit field index.
//
// Returns true if bit state was changed.
//
// Fails if bitindex is out of range.
func(st *State) ResetFlag(bitIndex uint32) (bool, error) {
	if bitIndex + 1 > st.BitSize {
		return false, fmt.Errorf("bit index %v is out of range of bitfield size %v", bitIndex, st.BitSize)
	}
	r := getFlag(bitIndex, st.Flags)
	if !r {
		return false, nil
	}
	byteIndex := bitIndex / 8
	localBitIndex := bitIndex % 8
	b := st.Flags[byteIndex] 
	st.Flags[byteIndex] = b & (^(1 << localBitIndex))
	return true, nil
}

// GetFlag returns the state of the flag at the given bit field index.
//
// Fails if bit field index is out of range.
func(st *State) GetFlag(bitIndex uint32) (bool, error) {
	if bitIndex + 1 > st.BitSize {
		return false, fmt.Errorf("bit index %v is out of range of bitfield size %v", bitIndex, st.BitSize)
	}
	return getFlag(bitIndex, st.Flags), nil
}

// FlagBitSize reports the amount of bits available in the bit field index.
func(st *State) FlagBitSize() uint32 {
	return st.BitSize
}

// FlagBitSize reports the amount of bits available in the bit field index.
func(st *State) FlagByteSize() uint8 {
	return uint8(len(st.Flags))
}

// MatchFlag matches the current state of the given flag.
//
// The flag is specified given its bit index in the bit field.
//
// If invertMatch is set, a positive result will be returned if the flag is not set.
func(st *State) MatchFlag(sig uint32, invertMatch bool) (bool, error) {
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

// GetIndex scans a byte slice in same order as in storage, and returns the index of the first set bit.
//
// If the given byte slice is too small for the bit field bitsize, the check will terminate at end-of-data without error.
func(st *State) GetIndex(flags []byte) bool {
	var globalIndex uint32
	if st.BitSize == 0 {
		return false
	}
	if len(flags) == 0 {
		return false
	}
	var byteIndex uint8
	var localIndex uint8
	l := uint8(len(flags))
	var i uint32
	for i = 0; i < st.BitSize; i++ {
		testVal := flags[byteIndex] & (1 << localIndex)
		if (testVal & st.Flags[byteIndex]) > 0 {
			return true
		}
		globalIndex += 1
		if globalIndex % 8 == 0 {
			byteIndex += 1
			localIndex = 0
			if byteIndex > (l - 1) {
				return false				
			}
		} else {
			localIndex += 1
		}
	}
	return false
}

// Where returns the current active rendering symbol.
func(st *State) Where() (string, uint16) {
	if len(st.ExecPath) == 0 {
		return "", 0
	}
	l := len(st.ExecPath)
	return st.ExecPath[l-1], st.SizeIdx
}

// Next moves to the next sink page index.
func(st *State) Next() (uint16, error) {
	if len(st.ExecPath) == 0 {
		return 0, fmt.Errorf("state root node not yet defined")
	}
	st.SizeIdx += 1
	s, idx := st.Where()
	Logg.Debugf("next page", "location", s, "index", idx)
	st.Moves += 1
	return st.SizeIdx, nil
}

func(st *State) Same() {
	st.Moves += 1
}

// Previous moves to the next sink page index.
//
// Fails if try to move beyond index 0.
func(st *State) Previous() (uint16, error) {
	if len(st.ExecPath) == 0 {
		return 0, fmt.Errorf("state root node not yet defined")
	}
	if st.SizeIdx == 0 {
		return 0, &IndexError{} // ("already at first index")
	}
	st.SizeIdx -= 1
	s, idx := st.Where()
	Logg.Debugf("previous page", "location", s, "index", idx)
	st.Moves += 1
	return st.SizeIdx, nil
}

// Sides informs the caller which index page options will currently succeed.
//
// Two values are returned, for the "next" and "previous" options in that order. A false value means the option is not available in the current state.
func(st *State) Sides() (bool, bool) {
	if len(st.ExecPath) == 0 {
		return false, false
	}
	next := true
	Logg.Tracef("sides", "index", st.SizeIdx)
	if st.SizeIdx == 0 {
		return next, false	
	}
	return next, true
}

// Top returns true if currently at topmode node.
//
// Fails if first Down() was never called.
func(st *State) Top() (bool, error) {
	if len(st.ExecPath) == 0 {
		return false, fmt.Errorf("state root node not yet defined")
	}
	return len(st.ExecPath) == 1, nil
}

// Down adds the given symbol to the command stack.
//
// Clears mapping and sink.
func(st *State) Down(input string) error {
	st.ExecPath = append(st.ExecPath, input)
	st.SizeIdx = 0
	st.Moves += 1
	return nil
}

// Up removes the latest symbol to the command stack, and make the previous symbol current.
//
// Frees all symbols and associated values loaded at the previous stack level. Cache capacity is increased by the corresponding amount.
//
// Clears mapping and sink.
//
// Fails if called at top frame.
func(st *State) Up() (string, error) {
	l := len(st.ExecPath)
	if l == 0 {
		return "", fmt.Errorf("exit called beyond top frame")
	}
	Logg.Tracef("execpath before", "path", st.ExecPath)
	st.ExecPath = st.ExecPath[:l-1]
	sym := ""
	if len(st.ExecPath) > 0 {
		sym = st.ExecPath[len(st.ExecPath)-1]
	}
	st.SizeIdx = 0
	Logg.Tracef("execpath after", "path", st.ExecPath)
	st.Moves += 1
	return sym, nil
}

// Depth returns the current call stack depth.
func(st *State) Depth() uint8 {
	return uint8(len(st.ExecPath)-1)
}

// Appendcode adds the given bytecode to the end of the existing code.
func(st *State) AppendCode(b []byte) error {
	st.Code = append(st.Code, b...)
	Logg.Debugf("code changed (append)", "code", b)
	return nil
}

// SetCode replaces the current bytecode with the given bytecode.
func(st *State) SetCode(b []byte) {
	Logg.Debugf("code changed (set)", "code", b)
	st.Code = b
}

// Get the remaning cached bytecode
func(st *State) GetCode() ([]byte, error) {
	b := st.Code
	st.Code = []byte{}
	return b, nil
}

// GetInput gets the most recent client input.
func(st *State) GetInput() ([]byte, error) {
	if st.input == nil {
		return nil, fmt.Errorf("no input has been set")
	}
	return st.input, nil
}

// SetInput is used to record the latest client input.
func(st *State) SetInput(input []byte) error {
	l := len(input)
	if l > 255 {
		return fmt.Errorf("input size %v too large (limit %v)", l, 255)
	}
	st.input = input
	return nil
}

// Reset re-initializes the state to run from top node with accumulated client state.
func(st *State) Restart() error {
	st.resetBaseFlags()
	st.Moves = 0
	st.SizeIdx = 0
	st.input = []byte{}
	return nil
}

// SetLanguage validates and sets language according to the given ISO639 language code.
func(st *State) SetLanguage(code string) error {
	if code == "" {
		st.Language = nil
	}
	l, err := lang.LanguageFromCode(code)
	if err != nil {
		return err
	}
	st.Language = &l
	Logg.Infof("language set", "language", l)
	return nil
}

// String implements String interface
func(st State) String() string {
	var flags string
	if st.debug {
		flags = FlagDebugger.AsString(st.Flags, st.BitSize - 8)
	} else {
		flags = fmt.Sprintf("0x%x", st.Flags)
	}
	var lang string
	if st.Language == nil {
		lang = "(default)"
	} else {
		lang = fmt.Sprintf("%s", *st.Language)
	}
	return fmt.Sprintf("moves: %v idx: %v flags: %s path: %s lang: %s", st.Moves, st.SizeIdx, flags, strings.Join(st.ExecPath, "/"), lang)
}

// initializes all flags not in control of client.
func(st *State) resetBaseFlags() {
	st.Flags[0] = 0
}

