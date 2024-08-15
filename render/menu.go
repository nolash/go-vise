package render

import (
	"context"
	"fmt"

	"git.defalsify.org/vise.git/resource"
)

// BrowseError is raised when browsing outside the page range of a rendered node.
type BrowseError struct {
	Idx uint16
	PageCount uint16
}

// Error implements the Error interface.
func(err *BrowseError) Error() string {
	return fmt.Sprintf("index is out of bounds: %v", err.Idx)
}

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

// Menu renders menus. May be included in a Page object to render menus for pages.
type Menu struct {
	rs resource.Resource
	menu [][2]string // selector and title for menu items.
	browse BrowseConfig // browse definitions.
	pageCount uint16 // number of pages the menu should represent.
	canNext bool // availability flag for the "next" browse option.
	canPrevious bool // availability flag for the "previous" browse option.
	//outputSize uint16 // maximum size constraint for the menu.
	sink bool
	keep bool
}

func(m Menu) String() string {
	return fmt.Sprintf("pagecount: %v menusink: %v next: %v prev: %v", m.pageCount, m.sink, m.canNext, m.canPrevious)
}

// NewMenu creates a new Menu with an explicit page count.
func NewMenu() *Menu {
	return &Menu{
		keep: true,
	}
}

// WithBrowseConfig defines the criteria for page browsing.
func(m *Menu) WithPageCount(pageCount uint16) *Menu {
	m.pageCount = pageCount
	return m
}

func(m *Menu) WithPages() *Menu {
	if m.pageCount == 0 {
		m.pageCount = 1
	}
	return m
}

func(m *Menu) WithSink() *Menu {
	m.sink = true
	return m
}

func(m *Menu) WithDispose() *Menu {
	m.keep = false
	return m
}

func(m *Menu) WithResource(rs resource.Resource) *Menu {
	m.rs = rs
	return m
}

func(m Menu) IsSink() bool {
	return m.sink
}

// WithSize defines the maximum byte size of the rendered menu.
//func(m *Menu) WithOutputSize(outputSize uint16) *Menu {
//	m.outputSize = outputSize
//	return m
//}

// GetOutputSize returns the defined heuristic menu size.
//func(m *Menu) GetOutputSize() uint32 {
//	return uint32(m.outputSize)
//}

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

// ReservedSize returns the maximum render byte size of the menu.
//func(m *Menu) ReservedSize() uint16 {
//	return m.outputSize
//}

 // mainSize, prevsize, nextsize, nextsize+prevsize
func(m *Menu) Sizes(ctx context.Context) ([4]uint32, error) {
	var menuSizes [4]uint32
	cfg := m.GetBrowseConfig()
	tmpm := NewMenu().WithBrowseConfig(cfg)
	v, err := tmpm.Render(ctx, 0)
	if err != nil {
		return menuSizes, err
	}
	menuSizes[0] = uint32(len(v))
	tmpm = tmpm.WithPageCount(2)
	v, err = tmpm.Render(ctx, 0)
	if err != nil {
		return menuSizes, err
	}
	menuSizes[1] = uint32(len(v)) - menuSizes[0]
	v, err = tmpm.Render(ctx, 1)
	if err != nil {
		return menuSizes, err
	}
	menuSizes[2] = uint32(len(v)) - menuSizes[0]
	menuSizes[3] = menuSizes[1] + menuSizes[2]
	return menuSizes, nil
}

func(m *Menu) titleFor(ctx context.Context, title string) (string, error) {
	if m.rs == nil {
		return title, nil
	}
	r, err := m.rs.GetMenu(ctx, title)
	if err != nil {
		return title, err
	}
	return r, nil
}

// Render returns the full current state of the menu as a string.
//
// After this has been executed, the state of the menu will be empty.
func(m *Menu) Render(ctx context.Context, idx uint16) (string, error) {
	var menuCopy [][2]string
	if m.keep {
		for _, v := range m.menu {
			menuCopy = append(menuCopy, v)
		}
	}

	err := m.applyPage(idx)
	if err != nil {
		return "", err
	}

	r := ""
	for true {
		l := len(r)
		//choice, title, err := m.shiftMenu()
		title, choice, err := m.shiftMenu()
		if err != nil {
			break
		}
		if l > 0 {
			r += "\n"
		}
		title, err = m.titleFor(ctx, title)
		if err != nil {
			return "", err
		}
		r += fmt.Sprintf("%s:%s", choice, title)
	}
	if m.keep {
		m.menu = menuCopy
	}
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
		return &BrowseError{Idx: idx, PageCount: m.pageCount}
		//return fmt.Errorf("index %v out of bounds (%v)", idx, m.pageCount)
	}
	
	m.reset()

	if idx == m.pageCount - 1 {
		m.canNext = false
	}
	if idx == 0 {
		m.canPrevious = false
	}
	Logg.Debugf("applypage", "m", m, "idx", idx)

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

// prepare menu object for re-use.
func(m *Menu) reset() {
	if m.browse.NextAvailable {
		m.canNext = true
	}
	if m.browse.PreviousAvailable {
		m.canPrevious = true
	}
}

func(m *Menu) Reset() {
	m.menu = [][2]string{}
	m.sink = false
	m.reset()
}
