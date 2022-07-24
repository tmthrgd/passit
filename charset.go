package passit

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type charset struct {
	runes []rune
	count int
}

// FromCharset returns a Template factory that generates passwords of count runes
// length by joining random runes from template. It returns an error if the template
// is invalid.
func FromCharset(template string) (func(count int) Template, error) {
	runes := []rune(template)
	if len(runes) < 2 {
		return nil, errors.New("password: template too short")
	} else if len(runes) > maxUint32 {
		return nil, errors.New("password: template too long")
	} else if !utf8.ValidString(template) {
		return nil, errors.New("password: template contains invalid unicode rune")
	} else if idx := strings.IndexFunc(template, notAllowed); idx >= 0 {
		r, _ := utf8.DecodeRuneInString(template[idx:])
		return nil, fmt.Errorf("password: template contains prohibited rune %U", r)
	}

	seen := make(map[rune]struct{}, len(runes))
	for _, r := range runes {
		if _, dup := seen[r]; dup {
			return nil, errors.New("password: template contains duplicate rune")
		}
		seen[r] = struct{}{}
	}

	return func(count int) Template { return &charset{runes, count} }, nil
}

func (c *charset) Password(r io.Reader) (string, error) {
	if c.count <= 0 {
		return "", errors.New("password: count must be greater than zero")
	}

	runes := make([]rune, c.count)
	for i := range runes {
		idx, err := readUint32n(r, uint32(len(c.runes)))
		if err != nil {
			return "", err
		}

		runes[i] = c.runes[idx]
	}

	return string(runes), nil
}

type rangeTable struct {
	tab   *unicode.RangeTable
	runes int
	count int
}

// FromRangeTable returns a Template factory that generates passwords of count
// runes length by joining random runes from the given unicode.RangeTable. It
// returns an error if the table has zero allowed runes. The table will be filtered
// by internally allowed runes.
func FromRangeTable(tab *unicode.RangeTable) (func(count int) Template, error) {
	tab = intersectRangeTables(allowedRangeTableStride1, tab)
	runes := countTableRunes(tab)
	if runes == 0 {
		return nil, errors.New("password: unicode.RangeTable contains zero allowed runes")
	}

	return func(count int) Template { return &rangeTable{tab, runes, count} }, nil
}

func (rt *rangeTable) Password(r io.Reader) (string, error) {
	if rt.count <= 0 {
		return "", errors.New("password: count must be greater than zero")
	}

	maybeUnicodeReadByte(r)

	runes := make([]rune, rt.count)
	for i := range runes {
		v, err := readRune(r, rt.tab, rt.runes)
		if err != nil {
			return "", err
		}

		runes[i] = v
	}

	return string(runes), nil
}

type emoji struct{ count int }

// Emoji returns a Template that generates passwords contain count number of emoji
// from the Unicode emoji list.
func Emoji(count int) Template { return &emoji{count} }

func (e *emoji) Password(r io.Reader) (string, error) {
	// TODO(tmthrgd): Remove once we switch to Unicode 13.0.
	maybeUnicodeReadByte(r)

	emoji := make([]string, e.count)
	for i := range emoji {
		idx, err := readUint32n(r, uint32(len(unicodeEmoji)))
		if err != nil {
			return "", err
		}

		emoji[i] = unicodeEmoji[idx]
	}

	return strings.Join(emoji, ""), nil
}
