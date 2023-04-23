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

type FsResource struct {
	MenuResource
	Path string
	fns map[string]EntryFunc
}

func NewFsResource(path string) FsResource {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return FsResource{
		Path: absPath,
	}
}

func(fsr FsResource) GetTemplate(sym string, ctx context.Context) (string, error) {
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

func(fsr FsResource) getFunc(sym string, input []byte, ctx context.Context) (Result, error) {
	v := ctx.Value("Language")
	if v == nil {
		return fsr.getFuncNoCtx(sym, input, nil)
	}
	language := v.(lang.Language)
	return fsr.getFuncNoCtx(sym, input, &language)
}

func(fsr FsResource) getFuncNoCtx(sym string, input []byte, language *lang.Language) (Result, error) {
	fb := sym + ".txt"
	fp := path.Join(fsr.Path, fb)
	fpl := fp
	if language != nil {
		fpl += "_" + language.Code
	}
	Logg.Debugf("getfunc search dir", "dir", fsr.Path, "path", fp, "lang_path", fpl, "sym", sym, "language", language)
	r, err := ioutil.ReadFile(fp)
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
