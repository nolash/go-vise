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
	PutMenu(string, string) error // Add a menu item.
	ShiftMenu() (string, string, error) // Remove and return the first menu item in list.
	SetMenuBrowse(string, string, bool) error // Set menu browser display details.
	RenderTemplate(sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) // Render the given data map using the template of the symbol.
	RenderMenu() (string, error) // Render the current state of menu
	Render(sym string, values map[string]string, idx uint16, sizer *Sizer) (string, error) // Render full output.
	FuncFor(sym string) (EntryFunc, error) // Resolve symbol content point for.
}

type MenuResource struct {
	menu [][2]string
	next [2]string
	prev [2]string
	sinkValues []string
	codeFunc CodeFunc
	templateFunc TemplateFunc
	funcFunc FuncForFunc
}

func NewMenuResource() *MenuResource {
	return &MenuResource{}
}

func(m *MenuResource) WithCodeGetter(codeGetter CodeFunc) *MenuResource {
	m.codeFunc = codeGetter
	return m
}

func(m *MenuResource) WithEntryFuncGetter(entryFuncGetter FuncForFunc) *MenuResource {
	m.funcFunc = entryFuncGetter
	return m
}

func(m *MenuResource) WithTemplateGetter(templateGetter TemplateFunc) *MenuResource {
	m.templateFunc = templateGetter
	return m
}

// SetMenuBrowse defines the how pagination menu options should be displayed.
//
// The selector is the expected user input, and the title is the description string.
//
// If back is set, the option will be defined for returning to a previous page.
func(m *MenuResource) SetMenuBrowse(selector string, title string, back bool) error {
	entry := [2]string{selector, title}
	if back {
		m.prev = entry
	} else {
		m.next = entry
	}
	return nil
}

// PutMenu adds a menu option to the menu rendering.
func(m *MenuResource) PutMenu(selector string, title string) error {
	m.menu = append(m.menu, [2]string{selector, title})
	log.Printf("menu %v", m.menu)
	return nil
}

// PutMenu removes and returns the first of remaining menu options.
//
// Fails if menu is empty.
func(m *MenuResource) ShiftMenu() (string, string, error) {
	if len(m.menu) == 0 {
		return "", "", fmt.Errorf("menu is empty")
	}
	r := m.menu[0]
	m.menu = m.menu[1:]
	return r[0], r[1], nil
}

// RenderMenu returns the full current state of the menu as a string.
//
// After this has been executed, the state of the menu will be empty.
func(m *MenuResource) RenderMenu() (string, error) {
	r := ""
	for true {
		l := len(r)
		choice, title, err := m.ShiftMenu()
		if err != nil {
			break
		}
		if l > 0 {
			r += "\n"
		}
		r += fmt.Sprintf("%s:%s", choice, title)
	}
	return r, nil
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
	s, err = m.RenderMenu()
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
