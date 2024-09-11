package db

import (
	"bytes"
	"context"
	"testing"

	"git.defalsify.org/vise.git/lang"
)

func TestDbBase(t *testing.T) {
	store := NewDbBase()
	store.SetPrefix(DATATYPE_STATE)
	if store.Prefix() != DATATYPE_STATE {
		t.Fatalf("expected %d, got %d", DATATYPE_STATE, store.Prefix())
	}
	if !store.Safe() {
		t.Fatal("expected safe")
	}
	store.SetLock(DATATYPE_MENU, false)
	if store.Safe() {
		t.Fatal("expected unsafe")
	}
	store.SetPrefix(DATATYPE_TEMPLATE)
	if store.CheckPut() {
		t.Fatal("expected checkput false")
	}
	store.SetLock(DATATYPE_TEMPLATE, false)
	if !store.CheckPut() {
		t.Fatal("expected checkput true")
	}
}

func TestDbKeyLanguage(t *testing.T) {
	ctx := context.Background()
	store := NewDbBase()
	store.SetPrefix(DATATYPE_TEMPLATE)
	l, err := lang.LanguageFromCode("nor")
	if err != nil {
		t.Fatal(err)
	}
	store.SetLanguage(&l)
	k, err := store.ToKey(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	v := append([]byte{DATATYPE_TEMPLATE}, []byte("foo")...)
	if !bytes.Equal(k.Default, v) {
		t.Fatalf("expected %x, got %x", v, k.Default)
	}
	v = append(v, []byte("_nor")...)
	if !bytes.Equal(k.Translation, v) {
		t.Fatalf("expected %x, got %x", v, k.Translation)
	}
}

func TestDbKeyNALanguage(t *testing.T) {
	ctx := context.Background()
	store := NewDbBase()
	store.SetPrefix(DATATYPE_STATE)
	l, err := lang.LanguageFromCode("nor")
	if err != nil {
		t.Fatal(err)
	}
	store.SetLanguage(&l)
	k, err := store.ToKey(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	v := append([]byte{DATATYPE_STATE}, []byte("foo")...)
	if !bytes.Equal(k.Default, v) {
		t.Fatalf("expected %x, got %x", v, k.Default)
	}
	if len(k.Translation) != 0 {
		t.Fatalf("expected no translation key, got %x", k.Translation)
	}

}

func TestDbKeyNoLanguage(t *testing.T) {
	ctx := context.Background()
	store := NewDbBase()
	store.SetPrefix(DATATYPE_TEMPLATE)
	k, err := store.ToKey(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	v := append([]byte{DATATYPE_TEMPLATE}, []byte("foo")...)
	if !bytes.Equal(k.Default, v) {
		t.Fatalf("expected %x, got %x", v, k.Default)
	}
	if len(k.Translation) != 0 {
		t.Fatalf("expected no translation key, got %x", k.Translation)
	}

}
