package asm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"git.defalsify.org/festive/vm"
)


type Asm struct {
	Instructions []*Instruction `@@*`
}

type Display struct {
	Sym string `@Sym Whitespace`
	Val string `@Quote @Sym @Quote Whitespace`
}

func(d Display) String() string {
	return fmt.Sprintf("Display: %v %v", d.Sym, d.Val)
}

//type Sig struct {
//	Sym string `@Sym Whitespace`
//	Size uint32 `@Size Whitespace`
//	Val uint32 `@Size Whitespace`
//}
//
//func(s Sig) String() string {
//	return fmt.Sprintf("Sig: %v %v %v", s.Sym, s.Size, s.Val)
//}

type Single struct {
	One string `@Sym Whitespace`
}

func(s Single) String() string {
	return fmt.Sprintf("Single: %v", s.One)
}

type Double struct {
	One string `@Sym Whitespace`
	Two string `@Sym Whitespace`
}

func(d Double) String() string {
	return fmt.Sprintf("Double: %v %v", d.One, d.Two)
}

type Sized struct {
	Sym string `@Sym Whitespace`
	Size uint32 `@Size Whitespace`
	X uint32 `(@Size Whitespace)?`
}

func(s Sized) String() string {
	return fmt.Sprintf("Sized: %v %v", s.Sym, s.Size)
}

type Arg struct {
	ArgNone string "Whitespace?"
	ArgDisplay *Display `@@?`
	ArgSized *Sized `@@?`
	ArgSingle *Single `@@?`
	ArgDouble *Double `@@?`
}

func (a Arg) String() string {
	if a.ArgDisplay != nil {
		return fmt.Sprintf("%s", a.ArgDisplay)
	}
	if a.ArgSized != nil {
		return fmt.Sprintf("%s", a.ArgSized)
	}
	if a.ArgSingle != nil {
		return fmt.Sprintf("%s", a.ArgSingle)
	}
	if a.ArgDouble != nil {
		return fmt.Sprintf("%s", a.ArgDouble)
	}
	return ""
}

type Instruction struct {
	OpCode string `@Ident`
	OpArg Arg `@@`
	Comment string `Comment?`
}

func (i Instruction) String() string {
	return fmt.Sprintf("%s %s", i.OpCode, i.OpArg)
}

var (
	asmLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Comment", `(?:#)[^\n]*\n?`},
		{"Ident", `^[A-Z]+`},
		{"Sym", `[a-zA-Z]+`},
		{"Size", `[0-9]+`},
		{"Whitespace", `[ \t\n\r]+`},
		{"Quote", `["']`},
	})
	asmParser = participle.MustBuild[Asm](
		participle.Lexer(asmLexer),
		participle.Elide("Comment", "Whitespace"),
	)
)

func numSize(n uint32) int {
	v := math.Log2(float64(n))
	return int(((v - 1) / 8) + 1)
}

func writeOpcode(op vm.Opcode, w *bytes.Buffer) (int, error) {
	bn := [2]byte{}
	binary.BigEndian.PutUint16(bn[:], uint16(op))
	n, err := w.Write(bn[:])
	return n, err
}

func writeSym(s string, w *bytes.Buffer) (int, error) {
	sz := len(s)
	if sz > 255 {
		return 0, fmt.Errorf("string size %v too big", sz)
	}
	w.Write([]byte{byte(sz)})
	return w.WriteString(s)
}

func writeSize(n uint32, w *bytes.Buffer) (int, error) {
	bn := [4]byte{}
	sz := numSize(n)
	if sz > 4 {
		return 0, fmt.Errorf("number size %v too big", sz)
	}
	w.Write([]byte{byte(sz)})
	binary.BigEndian.PutUint32(bn[:], n)
	c := 4-sz
	return w.Write(bn[c:])	
}


func parseSized(op vm.Opcode, arg Arg, w io.Writer) (int, error) {
	var rn int

	v := arg.ArgSized
	if v == nil {
		return 0, nil
	}

	b := bytes.NewBuffer(nil)

	n, err := writeOpcode(op, b)
	rn += n
	if  err != nil {
		return rn, err
	}
	
	n, err = writeSym(v.Sym, b)
	rn += n
	if err != nil {
		return rn, err
	}

	n, err = writeSize(v.Size, b)
	rn += n
	if err != nil {
		return rn, err
	}
	if w != nil {
		rn, err = w.Write(b.Bytes())
	} else {
		rn = 0
	}
	return rn, err
}

func Parse(s string, w io.Writer) (int, error) {
	rd := strings.NewReader(s)
	ast, err := asmParser.Parse("file", rd)
	var rn int

	for _, v := range ast.Instructions {
		op := vm.OpcodeIndex[v.OpCode]
		n, err := parseSized(op, v.OpArg, w)
		if err != nil {
			return n, err
		}
		if n > 0 {
			rn += n
			continue
		}
	}
	return rn, err
}
