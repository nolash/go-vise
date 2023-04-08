package asm

import (
	"fmt"

	"git.defalsify.org/festive/vm"
)

type BatchCode uint16

const (
	_MENU_OFFSET = 256
	MENU_DOWN = _MENU_OFFSET
	MENU_UP = _MENU_OFFSET + 1
	MENU_NEXT = _MENU_OFFSET + 2
	MENU_PREVIOUS = _MENU_OFFSET + 3
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
	if len(target) > 0 && bopCode != MENU_DOWN {
		return fmt.Errorf("target is only valid for DOWN")
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
		switch v.code {
		case MENU_UP:
			preLines = vm.NewLine(preLines, vm.MOUT, []string{v.choice, v.display}, nil, nil)
			postLines = vm.NewLine(postLines, vm.INCMP, []string{v.choice, "_"}, nil, nil)
		case MENU_NEXT:
			preLines = vm.NewLine(preLines, vm.MNEXT, []string{v.choice, v.display}, nil, nil)
			postLines = vm.NewLine(postLines, vm.INCMP, []string{v.choice, ">"}, nil, nil)
		case MENU_PREVIOUS:
			preLines = vm.NewLine(preLines, vm.MPREV, []string{v.choice, v.display}, nil, nil)
			postLines = vm.NewLine(postLines, vm.INCMP, []string{v.choice, "<"}, nil, nil)
		default:
			preLines = vm.NewLine(preLines, vm.MOUT, []string{v.choice, v.display}, nil, nil)
			postLines = vm.NewLine(postLines, vm.INCMP, []string{v.choice, v.target}, nil, nil)
		}
	}

	preLines = vm.NewLine(preLines, vm.HALT, nil, nil, nil)
	return append(preLines, postLines...)
}
