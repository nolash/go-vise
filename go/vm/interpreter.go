package vm

import (
	"fmt"
)

type BatchCode uint16

const (
	MENU_DOWN = 256
	MENU_UP = 257
	MENU_NEXT = 258
	MENU_PREVIOUS = 259
)

var (
	batchCode = map[string]BatchCode{
		"DOWN": MENU_DOWN,
		"UP": MENU_UP,
		"NEXT": MENU_NEXT,
		"PREVIOUS": MENU_PREVIOUS,
	}
)

type menuItem struct {
	code BatchCode
	choice string
	display string
	target string
}

type MenuProcessor struct {
	items []menuItem
	size uint32
}

func NewMenuProcessor() MenuProcessor {
	return MenuProcessor{}
}

func(mp *MenuProcessor) Add(bop string, choice string, display string, target string) error {
	bopCode := batchCode[bop]
	if bopCode == 0 {
		return fmt.Errorf("unknown menu instruction: %v", bop)
	}
	m := menuItem{
		code: bopCode,
		choice: choice,
		display: display,
		target: target,
	}
	mp.items = append(mp.items, m)
	return nil
}

func (mp *MenuProcessor) ToLines() []byte {
	preLines := []byte{}
	postLines := []byte{}

	for _, v := range mp.items {
		preLines = NewLine(preLines, MOUT, []string{v.choice, v.display}, nil, nil)
		switch v.code {
		case MENU_UP:
			postLines = NewLine(postLines, INCMP, []string{v.choice, "_"}, nil, nil)
		case MENU_NEXT:
			_ = postLines
		case MENU_PREVIOUS:
			_ = postLines
		default:
			postLines = NewLine(postLines, INCMP, []string{v.choice, v.target}, nil, nil)
		}
	}

	preLines = NewLine(preLines, HALT, nil, nil, nil)
	return append(preLines, postLines...)
}
