package resource

import (
	"context"
)

// EntryFunc is a function signature for retrieving value for a key
type EntryFunc func(ctx context.Context) (string, error)

// Resource implementation are responsible for retrieving values and templates for symbols, and can render templates from value dictionaries.
type Resource interface {
	GetTemplate(sym string) (string, error) // Get the template for a given symbol.
	GetCode(sym string) ([]byte, error) // Get the bytecode for the given symbol.
	RenderTemplate(sym string, values map[string]string) (string, error) // Render the given data map using the template of the symbol.
	FuncFor(sym string) (EntryFunc, error) // Resolve symbol code point for.
}
