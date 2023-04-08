package resource

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

type FsResource struct {
	MenuResource
	Path string
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
	return nil, fmt.Errorf("not implemented")
}

func(rs FsResource) String() string {
	return fmt.Sprintf("fs resource at path: %s", rs.Path)
}
