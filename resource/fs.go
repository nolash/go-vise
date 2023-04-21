package resource

import (
	"context"
	"fmt"
	"io/ioutil"
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

func NewFsResource(path string) (FsResource) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return FsResource{
		Path: absPath,
	}
}

func(fs FsResource) GetTemplate(sym string, ctx context.Context) (string, error) {
	fp := path.Join(fs.Path, sym)
	r, err := ioutil.ReadFile(fp)
	s := string(r)
	return strings.TrimSpace(s), err
}

func(fs FsResource) GetCode(sym string) ([]byte, error) {
	fb := sym + ".bin"
	fp := path.Join(fs.Path, fb)
	return ioutil.ReadFile(fp)
}

func(fs *FsResource) AddLocalFunc(sym string, fn EntryFunc) {
	if fs.fns == nil {
		fs.fns = make(map[string]EntryFunc)
	}
	fs.fns[sym] = fn
}

func(fs FsResource) FuncFor(sym string) (EntryFunc, error) {
	fn, ok := fs.fns[sym]
	if ok {
		return fn, nil
	}
	_, err := fs.getFuncNoCtx(sym, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unknown sym: %s", sym)
	}
	return fs.getFunc, nil
}

func(fs FsResource) String() string {
	return fmt.Sprintf("fs resource at path: %s", fs.Path)
}

func(fs FsResource) getFunc(sym string, input []byte, ctx context.Context) (Result, error) {
	v := ctx.Value("language")
	if v == nil {
		return fs.getFuncNoCtx(sym, input, nil)
	}
	language := v.(*lang.Language)
	return fs.getFuncNoCtx(sym, input, language)
}

func(fs FsResource) getFuncNoCtx(sym string, input []byte, language *lang.Language) (Result, error) {
	fb := sym + ".txt"
	fp := path.Join(fs.Path, fb)
	Logg.Debugf("getfunc search dir", "dir", fs.Path, "path", fp, "sym", sym, "language", language)
	r, err := ioutil.ReadFile(fp)
	if err != nil {
		return Result{}, fmt.Errorf("failed getting data for sym '%s': %v", sym, err)
	}
	s := string(r)
	return Result{
		Content: strings.TrimSpace(s),
	}, nil
}
