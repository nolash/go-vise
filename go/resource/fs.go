package resource

import (
	"context"
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

func(fs *FsResource) Get(sym string) (string, error) {
	return "", nil
}

func(fs *FsResource) Render(sym string, values []string) (string, error) {
	return "", nil
}
