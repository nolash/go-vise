package resource

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
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

func(fs FsResource) GetTemplate(sym string) (string, error) {
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

func(fs FsResource) FuncFor(sym string) (EntryFunc, error) {
	fn, ok := fs.fns[sym]
	if ok {
		return fn, nil
	}
	_, err := fs.getFuncNoCtx(sym)
	if err == nil {
		return nil, fmt.Errorf("unknown sym: %s", sym)
	}
	return fs.getFunc, nil
}

func(fs FsResource) String() string {
	return fmt.Sprintf("fs resource at path: %s", fs.Path)
}

func(fs FsResource) getFunc(sym string, ctx context.Context) (string, error) {
	return fs.getFuncNoCtx(sym)
}

func(fs FsResource) getFuncNoCtx(sym string) (string, error) {
	fb := sym + ".txt"
	fp := path.Join(fs.Path, fb)
	r, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", fmt.Errorf("failed getting data for sym '%s': %v", sym, err)
	}
	s := string(r)
	return strings.TrimSpace(s), err
}
