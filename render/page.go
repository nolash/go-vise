package render

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/resource"
)

// Page exectues output rendering into pages constrained by size.
type Page struct {
	cacheMap map[string]string // Mapped content symbols
	cache cache.Memory // Content store.
	resource resource.Resource // Symbol resolver.
	menu *Menu // Menu rendererer.
	sink *string // Content symbol rendered by dynamic size.
	sizer *Sizer // Process size constraints.
	err error // Error state to prepend to output.
}

// NewPage creates a new Page object.
func NewPage(cache cache.Memory, rs resource.Resource) *Page {
	return &Page{
		cache: cache,
		cacheMap: make(map[string]string),
		resource: rs,
	}
}

// WithMenu sets a menu renderer for the page.
func(pg *Page) WithMenu(menu *Menu) *Page {
	pg.menu = menu
	if pg.sizer != nil {
		pg.sizer = pg.sizer.WithMenuSize(pg.menu.ReservedSize())
	}
	return pg
}

// WithSizer sets a size constraints definition for the page.
func(pg *Page) WithSizer(sizer *Sizer) *Page {
	pg.sizer = sizer
	if pg.menu != nil {
		pg.sizer = pg.sizer.WithMenuSize(pg.menu.ReservedSize())
	}
	return pg
}

// WithError adds an error to prepend to the page output.
func(pg *Page) WithError(err error) *Page {
	pg.err = err
	return pg
}

// Error implements error interface.
func(pg *Page) Error() string {
	if pg.err != nil {
		return pg.err.Error()
	}
	return ""
}

// Usage returns size used by values and menu, and remaining size available
func(pg *Page) Usage() (uint32, uint32, error) {
	var l int
	var c uint16
	for k, v := range pg.cacheMap {
		l += len(v)
		sz, err := pg.cache.ReservedSize(k)
		if err != nil {
			return 0, 0, err
		}
		c += sz
	}
	r := uint32(l)
	rsv := uint32(c)-r
	if pg.menu != nil {
		r += uint32(pg.menu.ReservedSize())
	}
	return r, rsv, nil
}

// Map marks the given key for retrieval.
//
// After this, Val() will return the value for the key, and Size() will include the value size and limitations in its calculations.
//
// Only one symbol with no size limitation may be mapped at the current level.
func(pg *Page) Map(key string) error {
	v, err := pg.cache.Get(key)
	if err != nil {
		return err
	}
	l, err := pg.cache.ReservedSize(key)
	if err != nil {
		return err
	}
	if l == 0 {
		if pg.sink != nil && *pg.sink != key {
			return fmt.Errorf("sink already set to symbol '%v'", *pg.sink)
		}
		pg.sink = &key
	}
	pg.cacheMap[key] = v
	if pg.sizer != nil {
		err := pg.sizer.Set(key, l)
		if err != nil {
			return err
		}
	}
	Logg.Tracef("mapped", "key", key)
	return nil
}

// Val gets the mapped content for the given symbol.
//
// Fails if key is not mapped.
func(pg *Page) Val(key string) (string, error) {
	r := pg.cacheMap[key]
	if len(r) == 0 {
		return "", fmt.Errorf("key %v not mapped", key)
	}
	return r, nil
}

// Sizes returned the actual used bytes by each mapped symbol.
func(pg *Page) Sizes() (map[string]uint16, error) {
	sizes := make(map[string]uint16)
	var haveSink bool
	for k, _ := range pg.cacheMap {
		l, err := pg.cache.ReservedSize(k)
		if err != nil {
			return nil, err
		}
		if l == 0 {
			if haveSink {
				panic(fmt.Sprintf("duplicate sink for %v", k))
			}
			haveSink = true
		}
	}
	return sizes, nil
}

// RenderTemplate is an adapter to implement the builtin golang text template renderer as resource.RenderTemplate.
func(pg *Page) RenderTemplate(sym string, values map[string]string, idx uint16, ctx context.Context) (string, error) {
	tpl, err := pg.resource.GetTemplate(sym, ctx)
	if err != nil {
		return "", err
	}
	if pg.err != nil {
		derr := pg.Error()
		Logg.DebugCtxf(ctx, "prepending error", "err", pg.err, "display", derr)
		if len(tpl) == 0 {
			tpl = derr
		} else {
			tpl = fmt.Sprintf("%s\n%s", derr, tpl)
		}
	}
	if pg.sizer != nil {
		values, err = pg.sizer.GetAt(values, idx)
		if err != nil {
			return "", err
		}
	} else if idx > 0 {
		return "", fmt.Errorf("sizer needed for indexed render")
	}
	Logg.Debugf("render for", "index", idx)
	
	tp, err := template.New("tester").Option("missingkey=error").Parse(tpl)
	if err != nil {
		return "", err
	}


	b := bytes.NewBuffer([]byte{})
	err = tp.Execute(b, values)
	if err != nil {
		return "", err
	}
	return b.String(), err
}

// Render renders the current mapped content and menu state against the template associated with the symbol.
func(pg *Page) Render(sym string, idx uint16, ctx context.Context) (string, error) {
	var err error

	values, err := pg.prepare(sym, pg.cacheMap, idx, ctx)
	if err != nil {
		return "", err
	}

	return pg.render(sym, values, idx, ctx)
}

// Reset prepared the Page object for re-use.
//
// It clears mappings and removes the sink definition.
func(pg *Page) Reset() {
	pg.sink = nil
	pg.cacheMap = make(map[string]string)
	if pg.menu != nil {
		pg.menu.Reset()
	}
}


// render menu and all syms except sink, split sink into display chunks
// TODO: Function too long, split up
func(pg *Page) prepare(sym string, values map[string]string, idx uint16, ctx context.Context) (map[string]string, error) {
	var sink string

	if pg.sizer == nil {
		return values, nil
	}

	var sinkValues []string
	noSinkValues := make(map[string]string)
	for k, v := range values {
		sz, err := pg.cache.ReservedSize(k)
		if err != nil {
			return nil, err
		}
		if sz == 0 {
			sink = k
			sinkValues = strings.Split(v, "\n")
			v = ""
			Logg.Infof("found sink", "sym", k, "fields", len(sinkValues))
		}
		noSinkValues[k] = v
	}
	
	if sink == "" {
		Logg.Tracef("no sink found", "sym", sym)
		return values, nil
	}

	pg.sizer.AddCursor(0)
	s, err := pg.render(sym, noSinkValues, 0, ctx)
	if err != nil {
		return nil, err
	}
	// remaining includes core menu
	remaining, ok := pg.sizer.Check(s)
	if !ok {
		return nil, fmt.Errorf("capacity exceeded")
	}

	var menuSizes [4]uint32 // mainSize, prevsize, nextsize, nextsize+prevsize
	if pg.menu != nil {
		cfg := pg.menu.GetBrowseConfig()
		tmpm := NewMenu().WithBrowseConfig(cfg)
		v, err := tmpm.Render(0)
		if err != nil {
			return nil, err
		}
		menuSizes[0] = uint32(len(v))
		tmpm = tmpm.WithPageCount(2)
		v, err = tmpm.Render(0)
		if err != nil {
			return nil, err
		}
		menuSizes[1] = uint32(len(v)) - menuSizes[0]
		v, err = tmpm.Render(1)
		if err != nil {
			return nil, err
		}
		menuSizes[2] = uint32(len(v)) - menuSizes[0]
		menuSizes[3] = menuSizes[1] + menuSizes[2]
	}

	Logg.Debugf("calculated pre-navigation allocation", "bytes", remaining)

	l := 0
	var count uint16
	tb := strings.Builder{}
	rb := strings.Builder{}

	netRemaining := remaining - 1
	if len(sinkValues) > 1 {
		netRemaining -= menuSizes[1] - 1
	}

	for i, v := range sinkValues {
		l += len(v)
		Logg.Tracef("processing sink", "idx", i, "value", v)
		if uint32(l) > netRemaining - 1 {
			if tb.Len() == 0 {
				return nil, fmt.Errorf("capacity insufficient for sink field %v", i)
			}
			rb.WriteString(tb.String())
			rb.WriteRune('\n')
			c := uint32(rb.Len())
			pg.sizer.AddCursor(c)
			tb.Reset()
			l = 0
			if count == 0 {
				netRemaining -= menuSizes[2]
			}
			count += 1
		}
		if tb.Len() > 0 {
			tb.WriteByte(byte(0x00))
			l += 1
		}
		tb.WriteString(v)
	}

	if tb.Len() > 0 {
		rb.WriteString(tb.String())
		count += 1
	}

	r := rb.String()
	r = strings.TrimRight(r, "\n")

	noSinkValues[sink] = r

	if pg.menu != nil {
		pg.menu = pg.menu.WithPageCount(count)
	}

	for i, v := range strings.Split(r, "\n") {
		Logg.Tracef("nosinkvalue", "idx", i, "value", v)
	}

	return noSinkValues, nil
}

// render template, menu (if it exists), and audit size constraint (if it exists).
func(pg *Page) render(sym string, values map[string]string, idx uint16, ctx context.Context) (string, error) {
	var ok bool
	r := ""
	s, err := pg.RenderTemplate(sym, values, idx, ctx)
	if err != nil {
		return "", err
	}
	Logg.Debugf("rendered template", "bytes", len(s))
	r += s

	if pg.menu != nil {
		s, err = pg.menu.Render(idx)
		if err != nil {
			return "", err
		}
		Logg.Debugf("rendered menu", "bytes", len(s))
		if len(s) > 0 {
			r += "\n" + s
		}
	}

	if pg.sizer != nil {
		_, ok = pg.sizer.Check(r)
		if !ok {
			return "", fmt.Errorf("limit exceeded: %v", pg.sizer)
		}
	}
	return r, nil
}
