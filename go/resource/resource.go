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
	PutMenu(string, string) error // Add a menu item.
	ShiftMenu() (string, string, error) // Remove and return the first menu item in list.
	SetMenuBrowse(string, string, bool) error // Set menu browser display details.
	RenderTemplate(sym string, values map[string]string) (string, error) // Render the given data map using the template of the symbol.
	RenderMenu() (string, error) // Render the current state of menu
	FuncFor(sym string) (EntryFunc, error) // Resolve symbol content point for.
}

type MenuResource struct {
	menu [][2]string
	next [2]string
	prev [2]string
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
