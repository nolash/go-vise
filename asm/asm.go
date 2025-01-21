package asm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"git.defalsify.org/vise.git/vm"
)

// Asm assembles bytecode from the vise assembly mini-language.
//
// TODO: Conceal from outside use
type Asm struct {
	Instructions []*Instruction `@@*`
}

// Arg holds all parsed argument elements of a single line of assembly code.
//
// TODO: Conceal from outside use
type Arg struct {
	Sym      *string `(@Sym Whitespace?)?`
	Size     *uint32 `(@Size Whitespace?)?`
	Flag     *uint8  `(@Size Whitespace?)?`
	Selector *string `(@Sym Whitespace?)?`
	Desc     *string `(@Sym Whitespace?)?`
	//Desc *string `(Quote ((@Sym | @Size) @Whitespace?)+ Quote Whitespace?)?`
}

// writes the parsed instruction bytes to output.
func flush(b *bytes.Buffer, w io.Writer) (int, error) {
	if w != nil {
		return w.Write(b.Bytes())
	}
	return 0, nil
}

func parseTwoSym(b *bytes.Buffer, arg Arg) (int, error) {
	var rn int

	var selector string
	var sym string
	if arg.Size != nil {
		selector = strconv.FormatUint(uint64(*arg.Size), 10)
		//sym = *arg.Selector
		sym = *arg.Sym
	} else if arg.Selector != nil {
		if *arg.Sym == "*" {
			sym = *arg.Selector
			selector = *arg.Sym
		} else {
			sym = *arg.Sym
			selector = *arg.Selector
		}
	}

	n, err := writeSym(b, sym)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = writeSym(b, selector)
	rn += n
	if err != nil {
		return rn, err
	}
	return rn, nil
}

func parseTwoSymReverse(b *bytes.Buffer, arg Arg) (int, error) {
	var rn int

	sym := *arg.Selector
	selector := *arg.Sym
	n, err := writeSym(b, selector)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = writeSym(b, sym)
	rn += n
	if err != nil {
		return rn, err
	}

	return rn, nil
}

func parseSig(b *bytes.Buffer, arg Arg) (int, error) {
	var rn int

	n, err := writeSym(b, *arg.Sym)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = writeSize(b, *arg.Size)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = b.Write([]byte{uint8(*arg.Flag)})
	rn += n
	if err != nil {
		return rn, err
	}

	return rn, nil
}

func parseSized(b *bytes.Buffer, arg Arg) (int, error) {
	var rn int

	n, err := writeSym(b, *arg.Sym)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = writeSize(b, *arg.Size)
	rn += n
	if err != nil {
		return rn, err
	}

	return rn, nil
}

func parseFlagged(b *bytes.Buffer, arg Arg) (int, error) {
	var rn int

	n, err := writeSize(b, *arg.Size)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = b.Write([]byte{uint8(*arg.Flag)})
	rn += n
	if err != nil {
		return rn, err
	}

	return rn, nil

}

func parseOne(op vm.Opcode, instruction *Instruction, w io.Writer) (int, error) {
	a := instruction.OpArg
	var n_buf int
	var n_out int

	b := bytes.NewBuffer(nil)

	n, err := writeOpcode(b, op)
	n_buf += n
	if err != nil {
		return n_out, err
	}

	// Catch
	if a.Selector != nil {
		log.Printf("have selector %v", instruction)
		var n int
		var err error
		if op == vm.MOUT {
			n, err = parseTwoSymReverse(b, a)
		} else {
			n, err = parseTwoSym(b, a)
		}
		n_buf += n
		if err != nil {
			return n_out, err
		}
		return flush(b, w)
	}

	// Catch CATCH, LOAD and twosyms with integer-as-string
	if a.Size != nil {
		log.Printf("have size %v (%v)", instruction, *a.Size)
		if a.Sym == nil {
			n, err := parseFlagged(b, a)
			n_buf += n
			if err != nil {
				return n_out, err
			}
		} else {
			if a.Flag != nil {
				n, err := parseSig(b, a)
				n_buf += n
				if err != nil {
					return n_out, err
				}
			} else if op == vm.LOAD {
				n, err := parseSized(b, a)
				n_buf += n
				if err != nil {
					return n_out, err
				}
			} else {
				n, err := parseTwoSym(b, a)
				n_buf += n
				if err != nil {
					return n_out, err
				}

			}
		}
		return flush(b, w)
	}

	// Catch HALT
	if a.Sym == nil {
		return flush(b, w)
	}

	n, err = writeSym(b, *a.Sym)
	n_buf += n
	return flush(b, w)
}

// String implements the String interface.
func (a Arg) String() string {
	s := "[Arg]"
	if a.Sym != nil {
		s += " Sym: " + *a.Sym
	}
	if a.Size != nil {
		s += fmt.Sprintf(" Size: %v", *a.Size)
	}
	if a.Flag != nil {
		s += fmt.Sprintf(" Flag: %v", *a.Flag)
	}
	if a.Selector != nil {
		s += " Selector: " + *a.Selector
	}
	if a.Desc != nil {
		s += " Description: " + *a.Desc
	}

	return fmt.Sprintf(s)
}

// Instruction represents one full line of assembly code.
//
// TODO: Conceal from outside use
type Instruction struct {
	OpCode  string `@Ident`
	OpArg   Arg    `(Whitespace @@)?`
	Comment string `Comment? EOL`
}

// String implements the String interface.
func (i Instruction) String() string {
	return fmt.Sprintf("%s %s", i.OpCode, i.OpArg)
}

var (
	asmLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Comment", `(?:#)[^\n]*`},
		{"Ident", `^[A-Z]+`},
		{"Size", `[0-9]+`},
		{"Sym", `[a-zA-Z_\*\.\^\<\>][a-zA-Z0-9_]*`},
		{"Whitespace", `[ \t]+`},
		{"EOL", `[\n\r]+`},
		{"Quote", `["']`},
	})
	asmParser = participle.MustBuild[Asm](
		participle.Lexer(asmLexer),
		participle.Elide("Comment", "Whitespace"),
	)
)

func numSize(n uint32) int {
	v := math.Log2(float64(n))
	return int((v / 8) + 1)
}

func writeOpcode(w *bytes.Buffer, op vm.Opcode) (int, error) {
	bn := [2]byte{}
	binary.BigEndian.PutUint16(bn[:], uint16(op))
	n, err := w.Write(bn[:])
	return n, err
}

func writeSym(w *bytes.Buffer, s string) (int, error) {
	sz := len(s)
	if sz > 255 {
		return 0, fmt.Errorf("string size %v too big", sz)
	}
	w.Write([]byte{byte(sz)})
	return w.WriteString(s)
}

func writeSize(w *bytes.Buffer, n uint32) (int, error) {
	if n == 0 {
		return w.Write([]byte{0x01, 0x00})
	}
	bn := [4]byte{}
	sz := numSize(n)
	if sz > 4 {
		return 0, fmt.Errorf("number size %v too big", sz)
	}
	w.Write([]byte{byte(sz)})
	binary.BigEndian.PutUint32(bn[:], n)
	c := 4 - sz
	return w.Write(bn[c:])
}

// Batcher handles assembly commands that generates multiple instructions, such as menu navigation commands.
type Batcher struct {
	menuProcessor MenuProcessor
	inMenu        bool
}

// NewBatcher creates a new Batcher objcet.
func NewBatcher(mp MenuProcessor) Batcher {
	return Batcher{
		menuProcessor: NewMenuProcessor(),
	}
}

// MenuExit generates the instructions for the batch and writes them to the given io.Writer.
func (bt *Batcher) MenuExit(w io.Writer) (int, error) {
	if !bt.inMenu {
		return 0, nil
	}
	bt.inMenu = false
	b := bt.menuProcessor.ToLines()
	return w.Write(b)
}

// MenuAdd adds a new menu instruction to the batcher.
func (bt *Batcher) MenuAdd(w io.Writer, code string, arg Arg) (int, error) {
	bt.inMenu = true
	var selector string
	var sym string
	var display string
	if arg.Desc != nil {
		sym = *arg.Sym
		display = *arg.Desc
		selector = *arg.Selector
	} else if arg.Size != nil {
		if arg.Sym != nil {
			sym = *arg.Sym
		}
		selector = strconv.FormatUint(uint64(*arg.Size), 10)
		display = *arg.Selector
	} else {
		selector = *arg.Sym
		display = *arg.Selector
	}
	log.Printf("menu processor add %v '%v' '%v' '%v'", code, selector, display, sym)
	err := bt.menuProcessor.Add(code, selector, display, sym)
	return 0, err
}

// Exit is a synonym for MenuExit
func (bt *Batcher) Exit(w io.Writer) (int, error) {
	return bt.MenuExit(w)
}

// Parse one or more lines of assembly code, and write assembled bytecode to the provided writer.
func Parse(s string, w io.Writer) (int, error) {
	rd := strings.NewReader(s)
	ast, err := asmParser.Parse("file", rd)
	if err != nil {
		return 0, err
	}

	batch := Batcher{}

	var rn int
	for _, v := range ast.Instructions {
		log.Printf("parsing line %v: %v", v.OpCode, v.OpArg)
		op, ok := vm.OpcodeIndex[v.OpCode]
		if !ok {
			n, err := batch.MenuAdd(w, v.OpCode, v.OpArg)
			rn += n
			if err != nil {
				return rn, err
			}
		} else {
			n, err := batch.MenuExit(w)
			if err != nil {
				return rn, err
			}
			rn += n
			n, err = parseOne(op, v, w)
			rn += n
			if err != nil {
				return rn, err
			}
			log.Printf("wrote %v bytes for %v", n, v.OpArg)
		}
	}
	n, err := batch.Exit(w)
	rn += n
	if err != nil {
		return rn, err
	}
	rn += n

	return rn, err
}
