package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
)

func Loop(en *Engine, startSym string, ctx context.Context, reader io.Reader, writer io.Writer) error {
	err := en.Init(startSym, ctx)
	if err != nil {
		return fmt.Errorf("cannot init: %v\n", err)
	}

	b := bytes.NewBuffer(nil)
	en.WriteResult(b)
	fmt.Println(b.String())

	running := true
	bufReader := bufio.NewReader(reader)
	for running {
		in, err := bufReader.ReadString('\n')
		if err == io.EOF {
			log.Printf("EOF found, that's all folks")
			return nil
		}
		if err != nil {
			return fmt.Errorf("cannot read input: %v\n", err)
		}
		in = strings.TrimSpace(in)
		running, err = en.Exec([]byte(in), ctx)
		if err != nil {
			return fmt.Errorf("unexpected termination: %v\n", err)
		}
		b := bytes.NewBuffer(nil)
		en.WriteResult(b)
		writer.Write(b.Bytes())
		writer.Write([]byte{0x0a})
	}
	return nil
}
