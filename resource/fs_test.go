package resource

import (
	"context"
	"os"
	"path"
	"testing"

	"git.defalsify.org/vise.git/lang"
)

func TestNewFs(t *testing.T) {
	n := NewFsResource("./testdata")
	_ = n
}

func TestResourceLanguage(t *testing.T) {
	lang, err := lang.LanguageFromCode("nor")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()

	dir, err := os.MkdirTemp("", "vise_fsresource")
	if err != nil {
		t.Fatal(err)
	}

	fp := path.Join(dir, "foo")
	tmpl := "one two three"
	err = os.WriteFile(fp, []byte(tmpl), 0600)
	if err != nil {
		t.Fatal(err)
	}

	rs := NewFsResource(dir)
	r, err := rs.GetTemplate("foo", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if r != tmpl {
		t.Fatalf("expected '%s', got '%s'", tmpl, r)
	}

	ctx = context.WithValue(ctx, "Language", lang)
	rs = NewFsResource(dir)
	r, err = rs.GetTemplate("foo", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if r != tmpl {
		t.Fatalf("expected '%s', got '%s'", tmpl, r)
	}

	tmpl = "en to tre"
	err = os.WriteFile(fp + "_nor", []byte(tmpl), 0600)
	if err != nil {
		t.Fatal(err)
	}
	r, err = rs.GetTemplate("foo", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if r != tmpl {
		t.Fatalf("expected '%s', got '%s'", tmpl, r)
	}
}
