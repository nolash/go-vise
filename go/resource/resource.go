package resource

import (
	"context"
	"fmt"
	"log"
)

// EntryFunc is a function signature for retrieving value for a key
type EntryFunc func(ctx context.Context) (string, error)

// Resource implementation are responsible for retrieving values and templates for symbols, and can render templates from value dictionaries.
type Resource interface {
	GetTemplate(sym string) (string, error) // Get the template for a given symbol.
	GetCode(sym string) ([]byte, error) // Get the bytecode for the given symbol.
	PutMenu(string, string) error // Add a menu item
	ShiftMenu() (string, string, error)
	RenderTemplate(sym string, values map[string]string) (string, error) // Render the given data map using the template of the symbol.
	RenderMenu() (string, error)
	FuncFor(sym string) (EntryFunc, error) // Resolve symbol code point for.
}

type MenuResource struct {
	menu [][2]string
}

func(m *MenuResource) PutMenu(selector string, title string) error {
	m.menu = append(m.menu, [2]string{selector, title})
	log.Printf("menu %v", m.menu)
	return nil
}

func(m *MenuResource) ShiftMenu() (string, string, error) {
	if len(m.menu) == 0 {
		return "", "", fmt.Errorf("menu is empty")
	}
	r := m.menu[0]
	m.menu = m.menu[1:]
	return r[0], r[1], nil
}

func(m *MenuResource) RenderMenu() (string, error) {
	r := ""
	for true {
		l := len(r)
		choice, title, err := m.ShiftMenu()
		if err != nil {
			//if l == 0 { // TODO: replace with EOF
			//	return "", err
			//}
			break
		}
		if l > 0 {
			r += "\n"
		}
		r += fmt.Sprintf("%s:%s", choice, title)
	}
	return r, nil
}
