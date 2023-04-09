package render

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"git.defalsify.org/festive/cache"
	"git.defalsify.org/festive/resource"
)

type Page struct {
	cacheMap map[string]string // Mapped
	cache *cache.Cache
	resource resource.Resource
	menu *Menu
	sink *string
	sinkSize uint16
	sizer *Sizer
	sinkProcessed bool
}

func NewPage(cache *cache.Cache, rs resource.Resource) *Page {
	return &Page{
		cache: cache,
		cacheMap: make(map[string]string),
		resource: rs,
	}
}

func(pg *Page) WithMenu(menu *Menu) *Page {
	pg.menu = menu
	if pg.sizer != nil {
		pg.sizer = pg.sizer.WithMenuSize(pg.menu.ReservedSize())
	}
	return pg
}

func(pg *Page) WithSizer(sizer *Sizer) *Page {
	pg.sizer = sizer
	if pg.menu != nil {
		pg.sizer = pg.sizer.WithMenuSize(pg.menu.ReservedSize())
	}
	return pg
}

// Size returns size used by values and menu, and remaining size available
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
	if pg.menu != nil {
		r += uint32(pg.menu.ReservedSize())
	}
	return r, uint32(c)-r, nil
}

// Map marks the given key for retrieval.
//
// After this, Val() will return the value for the key, and Size() will include the value size and limitations in its calculations.
//
// Only one symbol with no size limitation may be mapped at the current level.
func(pg *Page) Map(key string) error {
	m, err := pg.cache.Get()
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
	pg.cacheMap[key] = m[key]
	if pg.sizer != nil {
		err := pg.sizer.Set(key, l)
		if err != nil {
			return err
		}
	}
	return nil
}

// Fails if key is not mapped.
func(pg *Page) Val(key string) (string, error) {
	r := pg.cacheMap[key]
	if len(r) == 0 {
		return "", fmt.Errorf("key %v not mapped", key)
	}
	return r, nil
}

// Moved from cache, MAP should hook to this object
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
		pg.sinkSize = l
	}
	return sizes, nil
}

// DefaultRenderTemplate is an adapter to implement the builtin golang text template renderer as resource.RenderTemplate.
func(pg *Page) RenderTemplate(sym string, values map[string]string, idx uint16) (string, error) {
	tpl, err := pg.resource.GetTemplate(sym)
	if err != nil {
		return "", err
	}
	if pg.sizer != nil {
		values, err = pg.sizer.GetAt(values, idx)
		if err != nil {
			return "", err
		}
	} else if idx > 0 {
		return "", fmt.Errorf("sizer needed for indexed render")
	}
	log.Printf("render for index: %v", idx)
	
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

// render menu and all syms except sink, split sink into display chunks
func(pg *Page) prepare(sym string, values map[string]string, idx uint16) (map[string]string, error) {
	var sink string
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
			log.Printf("found sink %s with field count %v", k, len(sinkValues))
		}
		noSinkValues[k] = v
	}
	
	if sink == "" {
		log.Printf("no sink found for sym %s", sym)
		return values, nil
	}

	pg.sizer.AddCursor(0)
	s, err := pg.render(sym, noSinkValues, 0)
	if err != nil {
		return nil, err
	}

	remaining, ok := pg.sizer.Check(s)
	if !ok {
		return nil, fmt.Errorf("capacity exceeded")
	}

	log.Printf("%v bytes available for sink split", remaining)

	l := 0
	var count uint16
	tb := strings.Builder{}
	rb := strings.Builder{}

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
			pg.sizer.AddCursor(c)
			tb.Reset()
			l = 0
			count += 1
		}
		if tb.Len() > 0 {
			tb.WriteByte(byte(0x00))
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
		log.Printf("nosinkvalue %v: %s", i, v)
	}

	return noSinkValues, nil
}

func(pg *Page) render(sym string, values map[string]string, idx uint16) (string, error) {
	var ok bool
	r := ""
	s, err := pg.RenderTemplate(sym, values, idx)
	if err != nil {
		return "", err
	}
	log.Printf("rendered %v bytes for template", len(s))
	r += s
	if pg.sizer != nil {
		_, ok = pg.sizer.Check(r)
		if !ok {
			return "", fmt.Errorf("limit exceeded: %v", pg.sizer)
		}
	}
	s, err = pg.menu.Render(idx)
	if err != nil {
		return "", err
	}
	log.Printf("rendered %v bytes for menu", len(s))
	r += "\n" + s
	if pg.sizer != nil {
		_, ok = pg.sizer.Check(r)
		if !ok {
			return "", fmt.Errorf("limit exceeded: %v", pg.sizer)
		}
	}
	return r, nil
}

func(pg *Page) Render(sym string, values map[string]string, idx uint16) (string, error) {
	var err error
	
	values, err = pg.prepare(sym, values, idx)
	if err != nil {
		return "", err
	}

	log.Printf("nosink %v", values)
	return pg.render(sym, values, idx)
}

func(pg *Page) Reset() {
	pg.sink = nil
	pg.sinkSize = 0
	pg.sinkProcessed = false
	pg.cacheMap = make(map[string]string)
}


