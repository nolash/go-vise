package router

import (
	"fmt"
)

type Router struct {
	selectors []string
	symbols map[string]string
}

func NewRouter() Router {
	return Router{
		symbols: make(map[string]string),
	}
}

func(r *Router) Add(selector string, symbol string) error {
	if r.symbols[selector] != "" {
		return fmt.Errorf("selector %v already set to symbol %v", selector, symbol)
	}
	l := len(selector)
	if (l > 255) {
		return fmt.Errorf("selector too long (is %v, max 255)", l)
	}
	l = len(symbol)
	if (l > 255) {
		return fmt.Errorf("symbol too long (is %v, max 255)", l)
	}
	r.selectors = append(r.selectors, selector)
	r.symbols[selector] = symbol
	return nil
}

func(r *Router) Next() []byte {
	if len(r.selectors) == 0 {
		return []byte{}
	}
	k := r.selectors[0]
	r.selectors = r.selectors[1:]
	v := r.symbols[k]
	if len(r.selectors) == 0 {
		r.symbols = nil	
	} else {
		delete(r.symbols, k)
	}
	lk := len(k)
	lv := len(v)
	b := []byte{uint8(lk)}
	b = append(b, k...)
	b = append(b, uint8(lv))	
	b = append(b, v...)
	return b
}

func(r *Router) ToBytes() []byte {
	b := []byte{}
	for true {
		v := r.Next()
		if len(v) == 0 {
			break
		}
		b = append(b, v...)
	}
	return b
}

func FromBytes(b []byte) Router {
	rb := NewRouter()
	for len(b) > 0 {
		l := b[0]
		k := b[1:1+l]
		b = b[1+l:]
		l = b[0]
		v := b[1:1+l]
		b = b[1+l:]
		rb.Add(string(k), string(v))
	}
	return rb
}
