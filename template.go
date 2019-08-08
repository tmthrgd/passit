package password

import (
	"io"
	"strings"
	"unicode"
)

type Template interface {
	Password(r io.Reader) (string, error)
}

var rangeTableASCII = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x20, Hi: 0x7e, Stride: 1},
	},
	LatinOffset: 1,
}

// TODO(tmthrgd): Review these ranges. PrintRanges is likely too permissive.
var allowedRanges = append(unicode.PrintRanges, rangeTableASCII)

func notAllowed(r rune) bool {
	if r <= 0x7e { // Fast path for ASCII.
		return r < 0x20
	}

	return !unicode.In(r, allowedRanges...)
}

type joined struct {
	ts []Template
}

func JoinTemplates(t ...Template) Template {
	if len(t) == 1 {
		return t[0]
	}

	return &joined{t}
}

func (j *joined) Password(r io.Reader) (string, error) {
	parts := make([]string, len(j.ts))
	for i, t := range j.ts {
		part, err := t.Password(r)
		if err != nil {
			return "", err
		}

		parts[i] = part
	}

	return strings.Join(parts, ""), nil
}

const Space = space
const space = fixedString(" ")

func FixedString(s string) Template { return fixedString(s) }

type fixedString string

func (s fixedString) Password(io.Reader) (string, error) { return string(s), nil }
