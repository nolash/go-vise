package resource

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// EntryFunc is a function signature for retrieving value for a key
type EntryFunc func(ctx context.Context) (string, error)
type CodeFunc func(sym string) ([]byte, error)
type TemplateFunc func(sym string, sizer *Sizer) (string, error)
type FuncForFunc func(sym string) (EntryFunc, error)

// Resource implementation are responsible for retrieving values and templates for symbols, and can render templates from value dictionaries.
type Resource interface {
	GetTemplate(sym string, sizer *Sizer) (string, error) // Get the template for a given symbol.
	GetCode(sym string) ([]byte, error) // Get the bytecode for the given symbol.
	RenderTemplate(sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) // Render the given data map using the template of the symbol.
	Render(sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) // Render full output.
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

// render menu and all syms except sink, split sink into display chunks
func(m *MenuResource) prepare(sym string, values map[string]string, idx uint16, sizer *Sizer) (map[string]string, error) {
	var sink string
	var sinkValues []string
	noSinkValues := make(map[string]string)
	for k, v := range values {
		sz, err := sizer.Size(k)
		if err != nil {
			return nil, err
		}
		if sz == 0 {
			sink = k
			sinkValues = strings.Split(v, "\n")
			v = ""
			log.Printf("found sink %s with field count %v", k, len(sinkValues))
		}
		noSinkValues[k] = v
	}
	
	if sink == "" {
		log.Printf("no sink found for sym %s", sym)
		return values, nil
	}

	s, err := m.render(sym, noSinkValues, 0, nil)
	if err != nil {
		return nil, err
	}

	remaining, ok := sizer.Check(s)
	if !ok {
		return nil, fmt.Errorf("capacity exceeded")
	}

	log.Printf("%v bytes available for sink split", remaining)

	l := 0
	tb := strings.Builder{}
	rb := strings.Builder{}

	sizer.AddCursor(0)
	for i, v := range sinkValues {
		log.Printf("processing sinkvalue %v: %s", i, v)
		l += len(v)
		if uint32(l) > remaining {
			if tb.Len() == 0 {
				return nil, fmt.Errorf("capacity insufficient for sink field %v", i)
			}
			rb.WriteString(tb.String())
			rb.WriteRune('\n')
			c := uint32(rb.Len())
			sizer.AddCursor(c)
			tb.Reset()
			l = 0
		}
		if tb.Len() > 0 {
			tb.WriteByte(byte(0x00))
		}
		tb.WriteString(v)
	}

	if tb.Len() > 0 {
		rb.WriteString(tb.String())
	}

	r := rb.String()
	r = strings.TrimRight(r, "\n")

	noSinkValues[sink] = r

	for i, v := range strings.Split(r, "\n") {
		log.Printf("nosinkvalue %v: %s", i, v)
	}

	return noSinkValues, nil
}

func(m *MenuResource) RenderTemplate(sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) {
	return DefaultRenderTemplate(m, sym, values, idx, sizer)
}

func(m *MenuResource) FuncFor(sym string) (EntryFunc, error) {
	return m.funcFunc(sym)
}

func(m *MenuResource) GetCode(sym string) ([]byte, error) {
	return m.codeFunc(sym)
}

func(m *MenuResource) GetTemplate(sym string, sizer *Sizer) (string, error) {
	return m.templateFunc(sym, sizer)
}

func(m *MenuResource) render(sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) {
	var ok bool
	r := ""
	s, err := m.RenderTemplate(sym, values, idx, sizer)
	if err != nil {
		return "", err
	}
	log.Printf("rendered %v bytes for template", len(s))
	r += s
	if sizer != nil {
		_, ok = sizer.Check(r)
		if !ok {
			return "", fmt.Errorf("limit exceeded: %v", sizer)
		}
	}
	s, err = m.RenderMenu(idx)
	if err != nil {
		return "", err
	}
	log.Printf("rendered %v bytes for menu", len(s))
	r += s
	if sizer != nil {
		_, ok = sizer.Check(r)
		if !ok {
			return "", fmt.Errorf("limit exceeded: %v", sizer)
		}
	}
	return r, nil
}

func(m *MenuResource) Render(sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) {
	var err error
	
	values, err = m.prepare(sym, values, idx, sizer)
	if err != nil {
		return "", err
	}

	return m.render(sym, values, idx, sizer)
}
