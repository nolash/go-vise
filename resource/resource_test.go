package resource

import (
	"context"
	"errors"
	"testing"
)

func codeGet(ctx context.Context, nodeSym string) ([]byte, error) {
	switch nodeSym {
		case "bar":
			return []byte("deafbeef"), nil
	}
	return nil, errors.New("unknown code")
}

func menuGet(ctx context.Context, menuSym string) (string, error) {
	switch menuSym {
		case "baz":
			return "xyzzy", nil
	}
	return "", errors.New("unknown code")

}

func templateGet(ctx context.Context, nodeSym string) (string, error) {
	switch nodeSym {
		case "tinkywinky":
			return "inky pinky {{.foo}} blinky clyde", nil
	}
	return "", errors.New("unknown code")
}

func entryFunc(ctx context.Context, nodeSym string, input []byte) (Result, error) {
	return Result{
		Content: "42",
	}, nil
}

func funcGet(ctx context.Context, loadSym string) (EntryFunc, error) {
	switch loadSym {
		case "dipsy":
			return entryFunc, nil
	}
	return nil, errors.New("unknown code")
}



func TestMenuResourceSetters(t *testing.T) {
	var err error
	ctx := context.Background()
	rs := NewMenuResource()
	_, err = rs.GetCode(ctx, "foo")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithCodeGetter(codeGet)
	_, err = rs.GetCode(ctx, "foo")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithCodeGetter(codeGet)
	_, err = rs.GetCode(ctx, "bar")
	if err != nil {
		t.Fatal(err)
	}

	_, err = rs.GetMenu(ctx, "bar")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithMenuGetter(menuGet)
	_, err = rs.GetMenu(ctx, "bar")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithMenuGetter(menuGet)
	_, err = rs.GetMenu(ctx, "baz")
	if err != nil {
		t.Fatal(err)
	}

	_, err = rs.GetTemplate(ctx, "baz")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithTemplateGetter(templateGet)
	_, err = rs.GetTemplate(ctx, "baz")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithTemplateGetter(templateGet)
	_, err = rs.GetTemplate(ctx, "tinkywinky")
	if err != nil {
		t.Fatal(err)
	}

	_, err = rs.FuncFor(ctx, "tinkywinky")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithEntryFuncGetter(funcGet)
	_, err = rs.FuncFor(ctx, "tinkywinky")
	if err == nil {
		errors.New("expected error")
	}

	rs.WithEntryFuncGetter(funcGet)
	_, err = rs.FuncFor(ctx, "dipsy")
	if err != nil {
		t.Fatal(err)
	}
}
