package passit

import (
	"errors"
	"io"
	"unicode"
	"unicode/utf8"
)

type asciiGenerator struct{ s string }

// Digit is a Generator that returns a random numeric digit.
var Digit Generator = &asciiGenerator{"0123456789"}

// LatinLower is a Generator that returns a random lowercase character from the latin
// alphabet.
var LatinLower Generator = &asciiGenerator{"abcdefghijklmnopqrstuvwxyz"}

// LatinUpper is a Generator that returns a random uppercase character from the latin
// alphabet.
var LatinUpper Generator = &asciiGenerator{"ABCDEFGHIJKLMNOPQRSTUVWXYZ"}

// LatinMixed is a Generator that returns a random mixed-case characters from the
// latin alphabet.
var LatinMixed Generator = &asciiGenerator{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"}

// LatinLower is a Generator that returns a random lowercase character from the latin
// alphabet or a numeric digit.
var LatinLowerDigit Generator = &asciiGenerator{"abcdefghijklmnopqrstuvwxyz0123456789"}

// LatinUpper is a Generator that returns a random uppercase character from the latin
// alphabet or a numeric digit.
var LatinUpperDigit Generator = &asciiGenerator{"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"}

// LatinMixed is a Generator that returns a random mixed-case characters from the
// latin alphabet or a numeric digit.
var LatinMixedDigit Generator = &asciiGenerator{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"}

func (ag *asciiGenerator) Password(r io.Reader) (string, error) {
	idx, err := readIntN(r, len(ag.s))
	if err != nil {
		return "", err
	}

	return ag.s[idx : idx+1], nil
}

type runeGenerator struct{ runes []rune }

// FromCharset returns a Generator that returns a random rune from charset. It
// returns an error if the charset is invalid.
func FromCharset(charset string) (Generator, error) {
	if !utf8.ValidString(charset) {
		return nil, errors.New("passit: charset contains invalid unicode rune")
	}

	runes := []rune(charset)
	if len(runes) > maxReadIntN {
		return nil, errors.New("passit: too many runes in charset")
	}

	switch len(runes) {
	case 0:
		return Empty, nil
	case 1:
		return FixedString(charset), nil
	default:
		return &runeGenerator{runes}, nil
	}
}

func (rg *runeGenerator) Password(r io.Reader) (string, error) {
	v, err := readSliceN(r, rg.runes)
	if err != nil {
		return "", err
	}

	return string(v), nil
}

type unicodeGenerator struct {
	tab   *unicode.RangeTable
	runes int
}

// FromRangeTable returns a Generator that returns a random rune from the
// unicode.RangeTable.
func FromRangeTable(tab *unicode.RangeTable) Generator {
	runes := countTableRunes(tab)
	if runes == 0 {
		return Empty
	}

	return &unicodeGenerator{tab, runes}
}

func (ug *unicodeGenerator) Password(r io.Reader) (string, error) {
	v, err := readRune(r, ug.tab, ug.runes)
	if err != nil {
		return "", err
	}

	return string(v), nil
}
