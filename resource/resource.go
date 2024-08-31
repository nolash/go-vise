package resource

import (
	"context"
	"fmt"
)


// Result contains the results of an external code operation.
type Result struct {
	Content string // content value for symbol after execution.
	Status int // application defined status code which can complement error returns
	FlagSet []uint32 // request caller to set error flags at given indices.
	FlagReset []uint32 // request caller to reset error flags at given indices.
}

// EntryFunc is a function signature for retrieving value for a key
type EntryFunc func(ctx context.Context, sym string, input []byte) (Result, error)
type CodeFunc func(ctx context.Context, sym string) ([]byte, error)
type MenuFunc func(ctx context.Context, sym string) (string, error)
type TemplateFunc func(ctx context.Context, sym string) (string, error)
type FuncForFunc func(sym string) (EntryFunc, error)

// Resource implementation are responsible for retrieving values and templates for symbols, and can render templates from value dictionaries.
type Resource interface {
	GetTemplate(ctx context.Context, sym string) (string, error) // Get the template for a given symbol.
	GetCode(ctx context.Context, sym string) ([]byte, error) // Get the bytecode for the given symbol.
	GetMenu(ctx context.Context, sym string) (string, error) // Receive menu test for menu symbol.
	FuncFor(sym string) (EntryFunc, error) // Resolve symbol content point for.
}

// MenuResource contains the base definition for building Resource implementations.
//
// TODO: Rename to BaseResource
type MenuResource struct {
	sinkValues []string
	codeFunc CodeFunc
	templateFunc TemplateFunc
	menuFunc MenuFunc
	funcFunc FuncForFunc
	fns map[string]EntryFunc
}

// NewMenuResource creates a new MenuResource instance.
func NewMenuResource() *MenuResource {
	rs := &MenuResource{}
	rs.funcFunc = rs.FallbackFunc
	return rs
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

// WithMenuGetter sets the menu symbol resolver method.
func(m *MenuResource) WithMenuGetter(menuGetter MenuFunc) *MenuResource {
	m.menuFunc = menuGetter
	return m
}

// FuncFor implements Resource interface.
func(m MenuResource) FuncFor(sym string) (EntryFunc, error) {
	return m.funcFunc(sym)
}

// GetCode implements Resource interface.
func(m MenuResource) GetCode(ctx context.Context, sym string) ([]byte, error) {
	return m.codeFunc(ctx, sym)
}

// GetTemplate implements Resource interface.
func(m MenuResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	return m.templateFunc(ctx, sym)
}

// GetMenu implements Resource interface.
func(m MenuResource) GetMenu(ctx context.Context, sym string) (string, error) {
	return m.menuFunc(ctx, sym)
}

// AddLocalFunc associates a handler function with a external function symbol to be returned by FallbackFunc.
func(m *MenuResource) AddLocalFunc(sym string, fn EntryFunc) {
	if m.fns == nil {
		m.fns = make(map[string]EntryFunc)
	}
	m.fns[sym] = fn
}

// FallbackFunc returns the default handler function for a given external function symbol.
func(m *MenuResource) FallbackFunc(sym string) (EntryFunc, error) {
	fn, ok := m.fns[sym]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", sym)
	}
	return fn, nil
}
