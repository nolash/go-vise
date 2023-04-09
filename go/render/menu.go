package render

import (
	"fmt"
)

// BrowseConfig defines the availability and display parameters for page browsing.
type BrowseConfig struct {
	NextAvailable bool
	NextSelector string
	NextTitle string
	PreviousAvailable bool
	PreviousSelector string
	PreviousTitle string
}

// Default browse settings for convenience.
func DefaultBrowseConfig() BrowseConfig {
	return BrowseConfig{
		NextAvailable: true,
		NextSelector: "11",
		NextTitle: "next",
		PreviousAvailable: true,
		PreviousSelector: "22",
		PreviousTitle: "previous",
	}
}

type Menu struct {
	menu [][2]string
	browse BrowseConfig
	pageCount uint16
	canNext bool
	canPrevious bool
	outputSize uint16
}

// NewMenu creates a new Menu with an explicit page count.
func NewMenu() *Menu {
	return &Menu{}
}

// WithBrowseConfig defines the criteria for page browsing.
func(m *Menu) WithPageCount(pageCount uint16) *Menu {
	m.pageCount = pageCount
	return m
}

// WithSize defines the maximum byte size of the rendered menu.
func(m *Menu) WithOutputSize(outputSize uint16) *Menu {
	m.outputSize = outputSize
	return m
}

// WithBrowseConfig defines the criteria for page browsing.
func(m *Menu) WithBrowseConfig(cfg BrowseConfig) *Menu {
	m.browse = cfg
	return m
}

// GetBrowseConfig returns a copy of the current state of the browse configuration.
func(m *Menu) GetBrowseConfig() BrowseConfig {
	return m.browse
}

// Put adds a menu option to the menu rendering.
func(m *Menu) Put(selector string, title string) error {
	m.menu = append(m.menu, [2]string{selector, title})
	return nil
}

// Render returns the full current state of the menu as a string.
//
// After this has been executed, the state of the menu will be empty.
func(m *Menu) Render(idx uint16) (string, error) {
	var menuCopy [][2]string
	for _, v := range m.menu {
		menuCopy = append(menuCopy, v)
	}

	err := m.applyPage(idx)
	if err != nil {
		return "", err
	}

	r := ""
	for true {
		l := len(r)
		choice, title, err := m.shiftMenu()
		if err != nil {
			break
		}
		if l > 0 {
			r += "\n"
		}
		r += fmt.Sprintf("%s:%s", choice, title)
	}
	m.menu = menuCopy
	return r, nil
}

// add available browse options.
func(m *Menu) applyPage(idx uint16) error {
	if m.pageCount == 0 {
		if idx > 0 {
			return fmt.Errorf("index %v > 0 for non-paged menu", idx)
		}
		return nil
	} else if idx >= m.pageCount {
		return fmt.Errorf("index %v out of bounds (%v)", idx, m.pageCount)
	}
	
	m.reset()

	if idx == m.pageCount - 1 {
		m.canNext = false
	}
	if idx == 0 {
		m.canPrevious = false
	}

	if m.canNext {
		err := m.Put(m.browse.NextSelector, m.browse.NextTitle)
		if err != nil {
			return err
		}
	}
	if m.canPrevious {
		err := m.Put(m.browse.PreviousSelector, m.browse.PreviousTitle)
		if err != nil {
			return err
		}
	}
	return nil
}

// removes and returns the first of remaining menu options.
// fails if menu is empty.
func(m *Menu) shiftMenu() (string, string, error) {
	if len(m.menu) == 0 {
		return "", "", fmt.Errorf("menu is empty")
	}
	r := m.menu[0]
	m.menu = m.menu[1:]
	return r[0], r[1], nil
}

func(m *Menu) reset() {
	if m.browse.NextAvailable {
		m.canNext = true
	}
	if m.browse.PreviousAvailable {
		m.canPrevious = true
	}
}

func(m *Menu) ReservedSize() uint16 {
	return m.outputSize
}
