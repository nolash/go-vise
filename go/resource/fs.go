package resource

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

type FsResource struct {
	path string
	ctx context.Context
}

func NewFsResource(path string, ctx context.Context) (FsResource) {
	return FsResource{
		path: path,
		ctx: ctx,
	}
}

func(fs FsResource) GetTemplate(sym string) (string, error) {
	fp := path.Join(fs.path, sym)
	r, err := ioutil.ReadFile(fp)
	s := string(r)
	return strings.TrimSpace(s), err
}

func(fs FsResource) RenderTemplate(sym string, values map[string]string) (string, error) {
	return "", nil
}

func(fs FsResource) GetCode(sym string) ([]byte, error) {
	return []byte{}, nil
}

func(fs FsResource) FuncFor(sym string) (EntryFunc, error) {
	return nil, fmt.Errorf("not implemented")
}
