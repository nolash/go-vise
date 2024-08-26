package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"git.defalsify.org/vise.git/asm"
)


type arg struct {
	One *string `@Sym`
	Two *string `((@Sym | @NumFirst) Whitespace?)?`
	Three *string `((@Sym | @NumFirst) Whitespace?)?`
	//Desc *string `(Quote ((@Sym | @Size) @Whitespace?)+ Quote Whitespace?)?`
}

type instruction struct {
	OpCode string `@Ident`
	OpArg arg `(Whitespace @@)?`
	Comment string `Comment? EOL`
}

type asmAsm struct {
	Instructions []*instruction `@@*`
}

func preProcess(b []byte) ([]byte, error) {
	asmLexer := lexer.MustSimple([]lexer.SimpleRule{
		{"Comment", `(?:#)[^\n]*`},
		{"Ident", `^[A-Z]+`},
		{"NumFirst", `[0-9][a-zA-Z0-9]*`},
		{"Sym", `[a-zA-Z_\*\.\^\<\>][a-zA-Z0-9_]*`},
		{"Whitespace", `[ \t]+`},
		{"EOL", `[\n\r]+`},
		{"Quote", `["']`},
	})
	asmParser := participle.MustBuild[asmAsm](
		participle.Lexer(asmLexer),
		participle.Elide("Comment", "Whitespace"),
	)
	ast, err := asmParser.ParseString("preprocessor", string(b))
	if err != nil {
		return nil, err
	}
	
	b = []byte{}
	for _, v := range ast.Instructions {
		s := []string{v.OpCode, *v.OpArg.One}
		if v.OpCode == "CATCH" {
			_, err := strconv.Atoi(*v.OpArg.Two)
			if err != nil {
				s = append(s, "42")	
			} else {
				s = append(s, *v.OpArg.Two)
			}
			s = append(s, *v.OpArg.Three)
		} else {
			for _, r := range []*string{v.OpArg.Two, v.OpArg.Three} {
				if r == nil {
					break
				}
				s = append(s, *r)
			}
		}
		b = append(b, []byte(strings.Join(s, " "))...)
		b = append(b, 0x0a)
	}

	return b, nil
}

func main() {
	if (len(os.Args) < 2) {
		os.Exit(1)
	}
	fp := os.Args[1]
	v, err := ioutil.ReadFile(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v", err)
		os.Exit(1)
	}
	log.Printf("start preprocessor")

	v, err = preProcess(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "preprocess error: %v", err)
		os.Exit(1)
	}
	log.Printf("preprocessor done")

	n, err := asm.Parse(string(v), os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v", err)
		os.Exit(1)
	}
	log.Printf("parsed total %v bytes", n)
}
