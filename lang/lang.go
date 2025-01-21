package lang

import (
	"context"
	"fmt"

	iso639_3 "github.com/barbashov/iso639-3"
)

var (
	// Default language (hard to get around)
	Default = "eng" // ISO639-3
)

// Language is used to set and get language translation to be used for rendering output.
type Language struct {
	Code string
	Name string
}

// LanguageFromCode returns a Language object from the given ISO-639-3 (three-letter) code.
//
// Will fail if an unknown code is provided.
func LanguageFromCode(code string) (Language, error) {
	r := iso639_3.FromAnyCode(code)
	if r == nil {
		return Language{}, fmt.Errorf("invalid language code: %s", code)
	}
	return Language{
		Code: r.Part3,
		Name: r.Name,
	}, nil
}

// String implements the String interface.
//
// Returns a representation of the Language fit for debugging.
func (l Language) String() string {
	return fmt.Sprintf("%s (%s)", l.Code, l.Name)
}

// LanguageFromContext scans the given context for a valid language
// specification, and if found returns a corresponding language
// object.
func LanguageFromContext(ctx context.Context) (Language, bool) {
	//var code string
	var lang Language
	var ok bool
	if ctx.Value("Language") != nil {
		lang = ctx.Value("Language").(Language)
		ok = true
		//code = lang.Code
	}
	return lang, ok
}
