package db

func TestDbBase(t *testing.T) {
	store := NewDbBase()
	store.SetPrefix(USERDATA_STATE)
	if !store.Prefix() == USERDATA_STATE {
		t.Fatal("expected %d, got %d", USERDATA_STATE, store.Prefix())
	}
	l, err := store.SetLanguage(lang.LanguageFromCode("nor"))
	if err != nil {
		t.Fatal(err)
	}
}
