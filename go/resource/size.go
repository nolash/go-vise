package resource

import (
	"fmt"
	"log"

	"git.defalsify.org/festive/state"
)

type Sizer struct {
	outputSize uint32
	menuSize uint16
	memberSizes map[string]uint16
	totalMemberSize uint32
	sink string
}

func SizerFromState(st *state.State) (Sizer, error){
	sz := Sizer{
		outputSize: st.GetOutputSize(),
		menuSize: st.GetMenuSize(),
		memberSizes: make(map[string]uint16),
	}
	sizes, err := st.Sizes()
	if err != nil {
		return sz, err
	}
	for k, v := range sizes {
		if v == 0 {
			sz.sink = k
		}
		sz.memberSizes[k] = v
		sz.totalMemberSize += uint32(v)
	}
	return sz, nil
}

func(szr *Sizer) Check(s string) (uint32, bool) {
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
	if szr.sink == s {
		return 0, nil
	}
	r, ok := szr.memberSizes[s]
	if !ok {
		return 0, fmt.Errorf("unknown member: %s", s)
	}
	return r, nil
}
