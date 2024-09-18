package engine

import (
	"context"
	"io"
)

// EngineIsh defines the interface for execution engines that handle vm initialization and execution, and rendering outputs.
type Engine interface {
	// Init sets the engine up for vm execution. It must be called before Exec.
	//Init(ctx context.Context) (bool, error)
	// Exec executes the pending bytecode.
	Exec(ctx context.Context, input []byte) (bool, error)
	// Flush renders output according to the state of VM execution
	// to the given io.Writer, and prepares the engine for another
	// VM execution.
	Flush(ctx context.Context, w io.Writer) (int, error)
	// Finish must be called after the last call to Exec.
	Finish() error
}
