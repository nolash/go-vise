package router

import (
	"fmt"
)

// Router contains and parses the routing section of the bytecode for a node.
type Router struct {
	selectors []string
	symbols map[string]string
}

// NewRouter creates a new Router object.
func NewRouter() Router {
	return Router{
		symbols: make(map[string]string),
	}
}

// NewStaticRouter creates a new Router object with a single destination.
//
// Used for routes that consume input value instead of navigation choices.
func NewStaticRouter(symbol string) Router {
	return Router{
		symbols: map[string]string{"_": symbol},
	}
}

// Add associates a selector with a destination symbol.
//
// Fails if:
// - selector or symbol value is invalid
// - selector already exists
func(r *Router) Add(selector string, symbol string) error {
	if r.symbols[selector] != "" {
		return fmt.Errorf("selector %v already set to symbol %v", selector, symbol)
	}
	l := len(selector)
	if (l > 255) {
		return fmt.Errorf("selector too long (is %v, max 255)", l)
	}
	if selector[0] == '_' {
		return fmt.Errorf("Invalid selector prefix '_'")
	}
	l = len(symbol)
	if (l > 255) {
		return fmt.Errorf("symbol too long (is %v, max 255)", l)
	}
	r.selectors = append(r.selectors, selector)
	r.symbols[selector] = symbol
	return nil
}

// Get retrieve symbol for selector.
//
// Returns an empty string if selector does not exist.
//
// Will always return an empty string if the router is static.
func(r *Router) Get(selector string) string {
	return r.symbols[selector]
}

// Get the statically defined symbol destination.
//
// Returns an empty string if not a static router.
func(r *Router) Default() string {
	return r.symbols["_"]
}

// Next removes one selector from the list of registered selectors.
//
// It returns it together with it associated value in bytecode form.
//
// Returns an empty byte array if no more selectors remain.
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

// ToBytes consume all selectors and values and returns them in sequence in bytecode form.
//
// This is identical to concatenating all returned values from non-empty Next() results.
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

// Restore a Router from bytecode.
//
// FromBytes(ToBytes()) creates an identical object.
func FromBytes(b []byte) Router {
	rb := NewRouter()
	navigable := true
	for len(b) > 0 {
		var k string
		l := b[0]
		if l == 0 {
			navigable = false
		} else {
			k = string(b[1:1+l])
			b = b[1+l:]
		}
		l = b[0]
		v := string(b[1:1+l])
		if !navigable {
			return NewStaticRouter(v)
		}
		b = b[1+l:]
		rb.Add(k, v)
	}
	return rb
}
