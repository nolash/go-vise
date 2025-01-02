package resource

import (
	"context"
	"testing"

	"git.defalsify.org/vise.git/testdata/testlocale"
	"git.defalsify.org/vise.git/lang"
)

func TestPoGetNotExist(t *testing.T) {
	ln, err := lang.LanguageFromCode("spa")
	if err != nil {
		t.Fatal(err)
	}

	rs := NewPoResource(ln, testlocale.LocaleDir)
	ctx := context.WithValue(context.Background(), "Language", ln)

	s, err := rs.GetMenu(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if s != "foo" {
		t.Fatalf("expected 'foo', got '%s'", s)
	}

	s, err = rs.GetTemplate(ctx, "bar")
	if err != nil {
		t.Fatal(err)
	}
	if s != "bar" {
		t.Fatalf("expected 'bar', got '%s'", s)
	}
}

func TestPoGet(t *testing.T) {
	ln, err := lang.LanguageFromCode("eng")
	if err != nil {
		t.Fatal(err)
	}
	rs := NewPoResource(ln, testlocale.LocaleDir)

	lnn, err := lang.LanguageFromCode("nor")
	if err != nil {
		t.Fatal(err)
	}
	rs = rs.WithLanguage(lnn)
	ctx := context.WithValue(context.Background(), "Language", lnn)

	s, err := rs.GetMenu(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if s != "fu" {
		t.Fatalf("expected 'fu', got '%s'", s)
	}

	s, err = rs.GetTemplate(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if s != "foo" {
		t.Fatalf("expected 'foo', got '%s'", s)
	}


	// eng now
	ctx = context.WithValue(context.Background(), "Language", ln)

	s, err = rs.GetMenu(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if s != "foo" {
		t.Fatalf("expected 'foo', got '%s'", s)
	}

	s, err = rs.GetMenu(ctx, "bar")
	if err != nil {
		t.Fatal(err)
	}
	if s != "bar" {
		t.Fatalf("expected 'bar', got '%s'", s)
	}

	s, err = rs.GetMenu(ctx, "inky")
	if err != nil {
		t.Fatal(err)
	}
	if s != "pinky" {
		t.Fatalf("expected 'pinky', got '%s'", s)
	}

	s, err = rs.GetTemplate(ctx, "bar")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf("expected 'baz', got '%s'", s)
	}
}
