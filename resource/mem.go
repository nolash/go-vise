package resource

import (
	"context"
	"fmt"
)

type MemResource struct {
	MenuResource
	templates map[string]string
	bytecodes map[string][]byte
	menus map[string]string
	funcs map[string]EntryFunc
}

func NewMemResource() MemResource {
	mr := MemResource{
		templates: make(map[string]string),
		bytecodes: make(map[string][]byte),
		funcs: make(map[string]EntryFunc),
	}
	mr.WithCodeGetter(mr.getCode)
	mr.WithTemplateGetter(mr.getTemplate)
	mr.WithEntryFuncGetter(mr.getFunc)
	mr.WithMenuGetter(mr.getMenu)
	return mr
}

func(mr MemResource) getTemplate(ctx context.Context, sym string) (string, error) {
	r, ok := mr.templates[sym]
	if !ok {
		return "", fmt.Errorf("unknown template symbol: %s", sym)
	}
	return r, nil
}

func(mr MemResource) getCode(sym string) ([]byte, error) {
	r, ok := mr.bytecodes[sym]
	if !ok {
		return nil, fmt.Errorf("unknown bytecode: %s", sym)
	}
	return r, nil
}

func(mr MemResource) getMenu(ctx context.Context, sym string) (string, error) {
	r, ok := mr.menus[sym]
	if !ok {
		r = sym
	}
	return r, nil

}

func(mr MemResource) getFunc(sym string) (EntryFunc, error) {
	r, ok := mr.funcs[sym]
	if !ok {
		return nil, fmt.Errorf("unknown entry func: %s", sym)
	}
	return r, nil
}

func(mr *MemResource) AddTemplate(sym string, tpl string) {
	Logg.Tracef("mem resource added template", "sym", sym, "length", len(tpl))
	mr.templates[sym] = tpl
}


func(mr *MemResource) AddEntryFunc(sym string, fn EntryFunc) {
	mr.funcs[sym] = fn
}

func(mr *MemResource) AddBytecode(sym string, code []byte) {
	Logg.Tracef("mem resource added bytecode", "sym", sym, "code", code)
	mr.bytecodes[sym] = code
}
