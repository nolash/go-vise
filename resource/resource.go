package resource

import (
	"context"
)

// Result contains the results of an external code operation.
type Result struct {
	Content string // content value for symbol after execution.
	FlagSet []uint32 // request caller to set error flags at given indices.
	FlagReset []uint32 // request caller to reset error flags at given indices.
}

// EntryFunc is a function signature for retrieving value for a key
type EntryFunc func(sym string, input []byte, ctx context.Context) (Result, error)
type CodeFunc func(sym string) ([]byte, error)
type TemplateFunc func(sym string, ctx context.Context) (string, error)
type FuncForFunc func(sym string) (EntryFunc, error)

// Resource implementation are responsible for retrieving values and templates for symbols, and can render templates from value dictionaries.
type Resource interface {
	GetTemplate(sym string, ctx context.Context) (string, error) // Get the template for a given symbol.
	GetCode(sym string) ([]byte, error) // Get the bytecode for the given symbol.
	FuncFor(sym string) (EntryFunc, error) // Resolve symbol content point for.
}

// MenuResource contains the base definition for building Resource implementations.
//
// TODO: Rename to BaseResource
type MenuResource struct {
	sinkValues []string
	codeFunc CodeFunc
	templateFunc TemplateFunc
	funcFunc FuncForFunc
}

// NewMenuResource creates a new MenuResource instance.
func NewMenuResource() *MenuResource {
	return &MenuResource{}
}

// WithCodeGetter sets the code symbol resolver method.
func(m *MenuResource) WithCodeGetter(codeGetter CodeFunc) *MenuResource {
	m.codeFunc = codeGetter
	return m
}

// WithEntryGetter sets the content symbol resolver getter method.
func(m *MenuResource) WithEntryFuncGetter(entryFuncGetter FuncForFunc) *MenuResource {
	m.funcFunc = entryFuncGetter
	return m
}

// WithTemplateGetter sets the template symbol resolver method.
func(m *MenuResource) WithTemplateGetter(templateGetter TemplateFunc) *MenuResource {
	m.templateFunc = templateGetter
	return m
}

// FuncFor implements Resource interface
func(m MenuResource) FuncFor(sym string) (EntryFunc, error) {
	return m.funcFunc(sym)
}

// GetCode implements Resource interface
func(m MenuResource) GetCode(sym string) ([]byte, error) {
	return m.codeFunc(sym)
}

// GetTemplate implements Resource interface
func(m MenuResource) GetTemplate(sym string, ctx context.Context) (string, error) {
	return m.templateFunc(sym, ctx)
}
