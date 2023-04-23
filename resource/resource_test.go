package resource

import (
	"context"
	"fmt"
)


type TestSizeResource struct {
	*MemResource
}

func NewTestSizeResource() TestSizeResource {
	mem := NewMemResource()
	rs := TestSizeResource{&mem}
	rs.AddTemplate("small", "one {{.foo}} two {{.bar}} three {{.baz}}")
	rs.AddTemplate("toobug", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus in mattis lorem. Aliquam erat volutpat. Ut vitae metus.")
	rs.AddTemplate("pages", "one {{.foo}} two {{.bar}} three {{.baz}}\n{{.xyzzy}}")
	rs.AddEntryFunc("foo", get)
	rs.AddEntryFunc("bar", get)
	rs.AddEntryFunc("baz", get)
	rs.AddEntryFunc("xyzzy", getXyzzy)
	return rs
}

func get(ctx context.Context, sym string, input []byte) (Result, error) {
	switch sym {
	case "foo":
		return Result{
			Content: "inky",
		}, nil
	case "bar":
		return Result{
			Content: "pinky",
		}, nil
	case "baz":
		return Result{
			Content: "blinky",
		}, nil
	}
	return Result{}, fmt.Errorf("unknown sym: %s", sym)
}

func getXyzzy(ctx context.Context, sym string, input []byte) (Result, error) {
	r := "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve"
	return Result{
		Content: r,
	}, nil
}
