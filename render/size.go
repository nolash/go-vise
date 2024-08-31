package render

import (
	"bytes"
	"fmt"
	"strings"
)

// Sizer splits dynamic contents into individual segments for browseable pages.
type Sizer struct {
	outputSize uint32 // maximum output for a single page.
//	menuSize uint16 // actual menu size for the dynamic page being sized
	memberSizes map[string]uint16 // individual byte sizes of all content to be rendered by template.
	totalMemberSize uint32 // total byte size of all content to be rendered by template (sum of memberSizes)
	crsrs []uint32 // byte offsets in the sink content for browseable pages indices.
	sink string // sink symbol.
}

// NewSizer creates a new Sizer object with the given output size constraint.
func NewSizer(outputSize uint32) *Sizer {
	return &Sizer{
		outputSize: outputSize,
		memberSizes: make(map[string]uint16),
	}
}

// WithMenuSize sets the size of the menu being used in the rendering context.
//func(szr *Sizer) WithMenuSize(menuSize uint16) *Sizer {
//	szr.menuSize = menuSize
//	return szr
//}

// Set adds a content symbol in the state it will be used by the renderer.
func(szr *Sizer) Set(key string, size uint16) error {
	szr.memberSizes[key] = size
	if size == 0 {
		szr.sink = key
	}
	szr.totalMemberSize += uint32(size)
	return nil
}

// Check audits whether the rendered string is within the output size constraint of the sizer.
func(szr *Sizer) Check(s string) (uint32, bool) {
	l := uint32(len(s))
	if szr.outputSize > 0 {
		if l > szr.outputSize {
			logg.Infof("sized check fails", "length", l, "sizer", szr)
			logg.Tracef("", "sizer contents", s)
			return 0, false
		}
		l = szr.outputSize - l
	}
	return l, true
}

// String implements the String interface.
//
// It outputs a representation of the Sizer fit for debug output.
func(szr *Sizer) String() string {
//	var diff uint32
//	if szr.outputSize > 0 {
//		diff = szr.outputSize - szr.totalMemberSize - uint32(szr.menuSize)
//	}
//	return fmt.Sprintf("output: %v, member: %v, menu: %v, diff: %v", szr.outputSize, szr.totalMemberSize, szr.menuSize, diff)
	return fmt.Sprintf("output: %v, member: %v", szr.outputSize, szr.totalMemberSize)
}

// Size gives the byte size of content for a single symbol.
//
// Fails if the symbol has not been registered using Set
func(szr *Sizer) Size(s string) (uint16, error) {
	r, ok := szr.memberSizes[s]
	if !ok {
		return 0, fmt.Errorf("unknown member: %s", s)
	}
	return r, nil
}

// Menusize returns the currently defined menu size.
//func(szr *Sizer) MenuSize() uint16 {
//	return szr.menuSize
//}

// AddCursor adds a pagination cursor for the paged sink content.
func(szr *Sizer) AddCursor(c uint32) {
	logg.Debugf("Added cursor", "offset", c)
	szr.crsrs = append(szr.crsrs, c)
}

// GetAt the paged symbols for the current page index.
//
// Fails if index requested is out of range.
func(szr *Sizer) GetAt(values map[string]string, idx uint16) (map[string]string, error) {
	if szr.sink == "" {
		return values, nil
	}
	outValues := make(map[string]string)
	for k, v := range values {
		logg.Tracef("check values", "k", k, "v", v, "idx", idx, "cursors", szr.crsrs)
		if szr.sink == k {
			if idx >= uint16(len(szr.crsrs)) {
				return nil, fmt.Errorf("no more values in index") 
			}
			c := szr.crsrs[idx]
			v = v[c:]
			nl := strings.Index(v, "\n")
			if nl > 0 {
				v = v[:nl]
			}
			b := bytes.ReplaceAll([]byte(v), []byte{0x00}, []byte{0x0a})
			v = string(b)
		}
		outValues[k] = v
	}
	return outValues, nil
}

// Reset flushes all size measurements, making the sizer available for reuse.
func(szr *Sizer) Reset() {
	szr.crsrs = []uint32{}
}
