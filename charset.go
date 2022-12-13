package passit

import (
	"errors"
	"io"
	"unicode"
	"unicode/utf8"
)

type asciiCharset struct{ s string }

// Number is a Template that returns a random numeric digit.
var Number Template = &asciiCharset{"0123456789"}

// LatinLower is a Template that returns a random lowercase character from the latin
// alphabet.
var LatinLower Template = &asciiCharset{"abcdefghijklmnopqrstuvwxyz"}

// LatinUpper is a Template that returns a random uppercase character from the latin
// alphabet.
var LatinUpper Template = &asciiCharset{"ABCDEFGHIJKLMNOPQRSTUVWXYZ"}

// LatinMixed is a Template that returns a random mixed-case characters from the
// latin alphabet.
var LatinMixed Template = &asciiCharset{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"}

// LatinLower is a Template that returns a random lowercase character from the latin
// alphabet.
var LatinLowerNumber Template = &asciiCharset{"abcdefghijklmnopqrstuvwxyz0123456789"}

// LatinUpper is a Template that returns a random uppercase character from the latin
// alphabet.
var LatinUpperNumber Template = &asciiCharset{"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"}

// LatinMixed is a Template that returns a random mixed-case characters from the
// latin alphabet.
var LatinMixedNumber Template = &asciiCharset{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"}

func (ac *asciiCharset) Password(r io.Reader) (string, error) {
	idx, err := readIntN(r, len(ac.s))
	if err != nil {
		return "", err
	}

	return ac.s[idx : idx+1], nil
}

type charset struct{ runes []rune }

// FromCharset returns a Template that returns a random rune from template. It
// returns an error if the template is invalid.
func FromCharset(template string) (Template, error) {
	if !utf8.ValidString(template) {
		return nil, errors.New("passit: template contains invalid unicode rune")
	}

	runes := []rune(template)
	if len(runes) > maxReadIntN {
		return nil, errors.New("passit: too many runes in template")
	}

	switch len(runes) {
	case 0:
		return Empty, nil
	case 1:
		return FixedString(template), nil
	default:
		return &charset{runes}, nil
	}
}

func (c *charset) Password(r io.Reader) (string, error) {
	idx, err := readIntN(r, len(c.runes))
	if err != nil {
		return "", err
	}

	return string(c.runes[idx : idx+1]), nil
}

type rangeTable struct {
	tab   *unicode.RangeTable
	runes int
}

// FromRangeTable returns a Template that returns a random rune from the
// unicode.RangeTable.
func FromRangeTable(tab *unicode.RangeTable) Template {
	runes := countTableRunes(tab)
	if runes == 0 {
		return Empty
	}

	return &rangeTable{tab, runes}
}

func (rt *rangeTable) Password(r io.Reader) (string, error) {
	v, err := readRune(r, rt.tab, rt.runes)
	if err != nil {
		return "", err
	}

	return string(v), nil
}
