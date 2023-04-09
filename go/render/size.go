package render

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

type Sizer struct {
	outputSize uint32
	menuSize uint16
	memberSizes map[string]uint16
	totalMemberSize uint32
	crsrs []uint32
	sink string
}

func NewSizer(outputSize uint32) *Sizer {
	return &Sizer{
		outputSize: outputSize,
		memberSizes: make(map[string]uint16),
	}
}

func(szr *Sizer) WithMenuSize(menuSize uint16) *Sizer {
	szr.menuSize = menuSize
	return szr
}

func(szr *Sizer) Set(key string, size uint16) error {
	szr.memberSizes[key] = size
	if size == 0 {
		szr.sink = key
	}
	szr.totalMemberSize += uint32(size)
	return nil
}

func(szr *Sizer) Check(s string) (uint32, bool) {
	log.Printf("sizercheck %s", s)
	l := uint32(len(s))
	if szr.outputSize > 0 {
		if l > szr.outputSize {
			log.Printf("sizer check fails with length %v: %s", l, szr)
			return 0, false
		}
		l = szr.outputSize - l
	}
	return l, true
}

func(szr *Sizer) String() string {
	var diff uint32
	if szr.outputSize > 0 {
		diff = szr.outputSize - szr.totalMemberSize - uint32(szr.menuSize)
	}
	return fmt.Sprintf("output: %v, member: %v, menu: %v, diff: %v", szr.outputSize, szr.totalMemberSize, szr.menuSize, diff)
}

func(szr *Sizer) Size(s string) (uint16, error) {
	r, ok := szr.memberSizes[s]
	if !ok {
		return 0, fmt.Errorf("unknown member: %s", s)
	}
	return r, nil
}

func(szr *Sizer) AddCursor(c uint32) {
	log.Printf("added cursor: %v", c)
	szr.crsrs = append(szr.crsrs, c)
}

func(szr *Sizer) GetAt(values map[string]string, idx uint16) (map[string]string, error) {
	if szr.sink == "" {
		return values, nil
	}
	outValues := make(map[string]string)
	for k, v := range values {
		if szr.sink == k {
			if idx >= uint16(len(szr.crsrs)) {
				return nil, fmt.Errorf("no more values in index") 
			}
			c := szr.crsrs[idx]
			v = v[c:]
			nl := strings.Index(v, "\n")
			log.Printf("k %v v %v c %v nl %v", k, v, c, nl)
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
