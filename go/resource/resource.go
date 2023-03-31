package resource

import (
	"context"
)

type EntryFunc func(input []byte, ctx context.Context) (string, error)

type Fetcher interface {
	Get(sym string) (string, error)
	Render(sym string, values map[string]string) (string, error)
	FuncFor(sym string) (EntryFunc, error)
}
