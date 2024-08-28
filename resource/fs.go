package resource

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.defalsify.org/vise.git/lang"
)

const (
	FSRESOURCETYPE_UNKNOWN = iota
	FSRESOURCETYPE_BIN
	FSRESOURCETYPE_TEMPLATE
)

type FsResource struct {
	MenuResource
	Path string
	fns map[string]EntryFunc
//	languageStrict bool
}

func NewFsResource(path string) *FsResource {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &FsResource{
		Path: absPath,
	}
}

//func(fsr *FsResource) WithStrictLanguage() *FsResource {
//	fsr.languageStrict = true
//	return fsr
//}

func(fsr FsResource) GetTemplate(ctx context.Context, sym string) (string, error) {
	fp := path.Join(fsr.Path, sym)
	fpl := fp
	v := ctx.Value("Language")
	if v != nil {
		lang := v.(lang.Language)
		fpl += "_" + lang.Code
	}
	var r []byte
	var err error
	r, err = ioutil.ReadFile(fpl)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if fpl != fp {
				r, err = ioutil.ReadFile(fp)
			}
		}
		if err != nil {
			return "", fmt.Errorf("failed getting template for sym '%s': %v", sym, err)
		}
	}
	s := string(r)
	return strings.TrimSpace(s), err
}

func(fsr FsResource) GetCode(sym string) ([]byte, error) {
	fb := sym + ".bin"
	fp := path.Join(fsr.Path, fb)
	return ioutil.ReadFile(fp)
}

func(fsr FsResource) GetMenu(ctx context.Context, sym string) (string, error) {
	fp := path.Join(fsr.Path, sym + "_menu")
	fpl := fp
	v := ctx.Value("Language")
	Logg.DebugCtxf(ctx, "getmenu", "lang", v, "path", fp)
	if v != nil {
		lang := v.(lang.Language)
		fpl += "_" + lang.Code
	}
	var r []byte
	var err error
	r, err = ioutil.ReadFile(fpl)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if fpl != fp {
				r, err = ioutil.ReadFile(fp)
			}
		}
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return sym, nil
			}
			return "", fmt.Errorf("failed getting template for sym '%s': %v", sym, err)
		}
	}
	s := string(r)
	return strings.TrimSpace(s), err
}

func(fsr *FsResource) AddLocalFunc(sym string, fn EntryFunc) {
	if fsr.fns == nil {
		fsr.fns = make(map[string]EntryFunc)
	}
	fsr.fns[sym] = fn
}

func(fsr FsResource) FuncFor(sym string) (EntryFunc, error) {
	fn, ok := fsr.fns[sym]
	if ok {
		return fn, nil
	}
	_, err := fsr.getFuncNoCtx(sym, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unknown sym: %s", sym)
	}
	return fsr.getFunc, nil
}

func(fsr FsResource) String() string {
	return fmt.Sprintf("fs resource at path: %s", fsr.Path)
}

func(fsr FsResource) getFunc(ctx context.Context, sym string, input []byte) (Result, error) {
	v := ctx.Value("Language")
	if v == nil {
		return fsr.getFuncNoCtx(sym, input, nil)
	}
	language := v.(lang.Language)
	return fsr.getFuncNoCtx(sym, input, &language)
}

func(fsr FsResource) getFuncNoCtx(sym string, input []byte, language *lang.Language) (Result, error) {
	fb := sym
	fbl := fb
	if language != nil {
		fbl += "_" + language.Code
	}
	fb += ".txt"
	fbl += ".txt"
	fp := path.Join(fsr.Path, fb)
	fpl := path.Join(fsr.Path, fbl)
	Logg.Debugf("getfunc search dir", "dir", fsr.Path, "path", fp, "lang_path", fpl, "sym", sym, "language", language)
	r, err := ioutil.ReadFile(fpl)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if fpl != fp {
				r, err = ioutil.ReadFile(fp)
			}
		}
		if err != nil {
			return Result{}, fmt.Errorf("failed getting data for sym '%s': %v", sym, err)
		}
	}
	s := string(r)
	return Result{
		Content: strings.TrimSpace(s),
	}, nil
}
