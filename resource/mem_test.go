package resource

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

func testEntry(ctx context.Context, sym string, input []byte) (Result, error) {
	return Result{
		Content: fmt.Sprintf("%sbar", input),
	}, nil
}

func TestMemResourceTemplate(t *testing.T) {
	rs := NewMemResource()
	rs.AddTemplate("foo", "bar")

	ctx := context.TODO()
	r, err := rs.GetTemplate(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if r != "bar" {
		fmt.Errorf("expected 'bar', got %s", r)
	}

	_, err = rs.GetTemplate(ctx, "bar")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestMemResourceCode(t *testing.T) {
	rs := NewMemResource()
	rs.AddBytecode("foo", []byte("bar"))

	r, err := rs.GetCode("foo")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(r, []byte("bar")) {
		fmt.Errorf("expected 'bar', got %x", r)
	}

	_, err = rs.GetCode("bar")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestMemResourceEntry(t *testing.T) {
	rs := NewMemResource()
	rs.AddEntryFunc("foo", testEntry)

	fn, err := rs.FuncFor("foo")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.TODO()
	r, err := fn(ctx, "foo", []byte("xyzzy"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Content != "foobar" {
		fmt.Errorf("expected 'foobar', got %x", r.Content)
	}

	_, err = rs.FuncFor("bar")
	if err == nil {
		t.Fatalf("expected error")
	}
}
