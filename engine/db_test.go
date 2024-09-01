package engine

import (
	"context"
	"fmt"
	"testing"

	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/vm"
)

func codeGet(ctx context.Context, s string) ([]byte, error) {
	var b []byte
	var err error
	switch s {
		case "root":
			b = vm.NewLine(nil, vm.HALT, nil, nil, nil)
			b = vm.NewLine(b, vm.LOAD, []string{"foo"}, []byte{0x0}, nil)
		default:
			err = fmt.Errorf("unknown code symbol '%s'", s)
	}
	return b, err
}

func TestDbEngineMinimal(t *testing.T) {
	ctx := context.Background()
	cfg := Config{}
	rs := resource.NewMenuResource()
	en := NewDbEngine(cfg, rs)
	cont, err := en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cont {
		t.Fatalf("expected not continue")
	}
}

func TestDbEngineRoot(t *testing.T) {
	ctx := context.Background()
	cfg := Config{}
	rs := resource.NewMenuResource()
	rs.WithCodeGetter(codeGet)
	en := NewDbEngine(cfg, rs)
	cont, err := en.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !cont {
		t.Fatalf("expected continue")
	}
}
