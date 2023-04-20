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
func Loop(en EngineIsh, reader io.Reader, writer io.Writer, ctx context.Context) error {
	defer en.Finish()
	var err error
	_, err = en.WriteResult(writer, ctx)
	if err != nil {
		return err
	}
	writer.Write([]byte{0x0a})

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
		running, err = en.Exec([]byte(in), ctx)
		if err != nil {
			return fmt.Errorf("unexpected termination: %v\n", err)
		}
		_, err = en.WriteResult(writer, ctx)
		if err != nil {
			return err
		}
		writer.Write([]byte{0x0a})

	}
	return nil
}
