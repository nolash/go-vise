package lang

import (
	"fmt"

	"github.com/barbashov/iso639-3"
)

var (
	Default = "eng" // ISO639-3
)

type Language struct {
	Code string
	Name string
}

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

func(l Language) String() string {
	return fmt.Sprintf("%s (%s)", l.Code, l.Name)
}
