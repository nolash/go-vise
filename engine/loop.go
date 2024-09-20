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
//
// If initial is set, the value will be used for the first (initializing) execution
// If nil, an empty byte value will be used.
func Loop(ctx context.Context, en Engine, reader io.Reader, writer io.Writer, initial []byte) error {
	defer en.Finish()
	if initial == nil {
		initial = []byte{}
	}
	cont, err := en.Exec(ctx, initial)
	if err != nil {
		return err
	}
	l, err := en.Flush(ctx, writer)
	if err != nil {
		if err != ErrFlushNoExec {
			return err
		}
	}
	if l > 0 {
		writer.Write([]byte{0x0a})
	}
	if !cont {
		return nil
	}

	running := true
	bufReader := bufio.NewReader(reader)
	for running {
		in, err := bufReader.ReadString('\n')
		if err == io.EOF {
			logg.DebugCtxf(ctx, "EOF found, that's all folks")
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
		l, err := en.Flush(ctx, writer)
		if err != nil {
			return err
		}
		if l > 0 {
			writer.Write([]byte{0x0a})
		}
	}
	return nil
}
