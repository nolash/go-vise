package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// Loop starts an engine execution loop with the given symbol as the starting node.
//
// The root reads inputs from the provided reader, one line at a time.
//
// It will execute until running out of bytecode in the buffer.
//
// Any error not handled by the engine will terminate the oop and return an error.
//
// Rendered output is written to the provided writer.
func Loop(ctx context.Context, en EngineIsh, reader io.Reader, writer io.Writer) error {
	defer en.Finish()
	l, err := en.WriteResult(ctx, writer)
	if err != nil {
		return err
	}
	if l > 0 {
		writer.Write([]byte{0x0a})
	}

	running := true
	bufReader := bufio.NewReader(reader)
	for running {
		in, err := bufReader.ReadString('\n')
		if err == io.EOF {
			Logg.DebugCtxf(ctx, "EOF found, that's all folks")
			return nil
		}
		if err != nil {
			return fmt.Errorf("cannot read input: %v\n", err)
		}
		in = strings.TrimSpace(in)
		running, err = en.Exec(ctx, []byte(in))
		if err != nil {
			return fmt.Errorf("unexpected termination: %v\n", err)
		}
		l, err := en.WriteResult(ctx, writer)
		if err != nil {
			return err
		}
		if l > 0 {
			writer.Write([]byte{0x0a})
		}
	}
	return nil
}
