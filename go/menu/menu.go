package menu

import (
	"fmt"
	"log"
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

// WithBrowseConfig defines the criteria for page browsing.
func(m *Menu) WithBrowseConfig(cfg BrowseConfig) *Menu {
	m.browse = cfg
	
	return m
}

// Put adds a menu option to the menu rendering.
func(m *Menu) Put(selector string, title string) error {
	m.menu = append(m.menu, [2]string{selector, title})
	log.Printf("menu %v", m.menu)
	return nil
}

// Render returns the full current state of the menu as a string.
//
// After this has been executed, the state of the menu will be empty.
func(m *Menu) Render(idx uint16) (string, error) {
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
	return r, nil
}

// add available browse options.
func(m *Menu) applyPage(idx uint16) error {
	if m.pageCount == 0 {
		return nil
	} else if idx >= m.pageCount {
		return fmt.Errorf("index %v out of bounds (%v)", idx, m.pageCount)
	}
	if m.browse.NextAvailable {
		m.canNext = true
	}
	if m.browse.PreviousAvailable {
		m.canPrevious = true
	}
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


