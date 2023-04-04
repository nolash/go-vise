package asm

import (
	"fmt"
	"io"
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

type Sig struct {
	Sym string `@Sym Whitespace`
	Size uint32 `@Size Whitespace`
	Val uint32 `@Size Whitespace`
}

func(s Sig) String() string {
	return fmt.Sprintf("Sig: %v %v %v", s.Sym, s.Size, s.Val)
}

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

func Parse(s string, w io.Writer) (int, error) {
	rd := strings.NewReader(s)
	ast, err := asmParser.Parse("file", rd)
	for i, v := range ast.Instructions {
		op := vm.OpcodeIndex[v.OpCode]
		fmt.Printf("%v (%v) %v\n", i, op, v)
	}
	return 0, err
}
