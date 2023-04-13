package resource

import (
	"context"
)

// EntryFunc is a function signature for retrieving value for a key
type EntryFunc func(sym string, input []byte, ctx context.Context) (string, error)
type CodeFunc func(sym string) ([]byte, error)
type TemplateFunc func(sym string) (string, error)
type FuncForFunc func(sym string) (EntryFunc, error)

// Resource implementation are responsible for retrieving values and templates for symbols, and can render templates from value dictionaries.
type Resource interface {
	GetTemplate(sym string) (string, error) // Get the template for a given symbol.
	GetCode(sym string) ([]byte, error) // Get the bytecode for the given symbol.
	FuncFor(sym string) (EntryFunc, error) // Resolve symbol content point for.
}

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

func(m *MenuResource) FuncFor(sym string) (EntryFunc, error) {
	return m.funcFunc(sym)
}

func(m *MenuResource) GetCode(sym string) ([]byte, error) {
	return m.codeFunc(sym)
}

func(m *MenuResource) GetTemplate(sym string) (string, error) {
	return m.templateFunc(sym)
}

