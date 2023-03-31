package resource

import (
	"context"
)

// EntryFunc is a function signature for retrieving value for a key
type EntryFunc func(ctx context.Context) (string, error)

// Resource implementation are responsible for retrieving values and templates for symbols, and can render templates from value dictionaries.
type Resource interface {
	Get(sym string) (string, error)
	Render(sym string, values map[string]string) (string, error)
	FuncFor(sym string) (EntryFunc, error)
}
