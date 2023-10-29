package passit

import (
	"io"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type transformGenerator struct {
	gen Generator
	fn  func(string) string
}

// Transform returns a Generator that converts the generated password according to
// the user supplied function fn.
func Transform(gen Generator, fn func(string) string) Generator {
	return &transformGenerator{gen, fn}
}

func (tg *transformGenerator) Password(r io.Reader) (string, error) {
	pass, err := tg.gen.Password(r)
	return tg.fn(pass), err
}

// LowerCase returns a Generator that uses [strings.ToLower] to map all Unicode
// letters in the generated password to their lower case.
func LowerCase(gen Generator) Generator {
	return Transform(gen, strings.ToLower)
}

// UpperCase returns a Generator that uses [strings.ToUpper] to map all Unicode
// letters in the generated password to their upper case.
func UpperCase(gen Generator) Generator {
	return Transform(gen, strings.ToUpper)
}

// TitleCase returns a Generator that uses [golang.org/x/text/cases.Title] to
// convert the generated password to language-specific title case.
func TitleCase(gen Generator, t language.Tag) Generator {
	return Transform(gen, cases.Title(t).String)
}
