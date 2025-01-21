package lang

import (
	"testing"
)

func TestLang(t *testing.T) {
	var err error
	_, err = LanguageFromCode("xxx")
	if err == nil {
		t.Fatalf("expected error")
	}
	l, err := LanguageFromCode("en")
	if err != nil {
		t.Fatal(err)
	}
	if l.Code != "eng" {
		t.Fatalf("expected 'eng', got '%s'", l.Code)
	}
}
