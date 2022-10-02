// Package passit provides various password generators.
package passit

import (
	"errors"
	"io"
	"strings"
)

// Template is an interface for generating passwords.
type Template interface {
	// Password returns a randomly generated password using r as the source of
	// randomness.
	//
	// The returned password may or may not be deterministic with respect to r.
	//
	// r should be a uniformly random stream. The numbers of bytes read from r
	// may exceed the number of characters in the returned password.
	Password(r io.Reader) (string, error)
}

type joined struct {
	ts  []Template
	sep string
}

// Join returns a Template that generates a password that is the concatenation of
// all the given Templates.
func Join(sep string, t ...Template) Template {
	switch len(t) {
	case 0:
		return FixedString("")
	case 1:
		return t[0]
	default:
		return &joined{append([]Template(nil), t...), sep}
	}
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

	return strings.Join(parts, j.sep), nil
}

type randomCount struct {
	tmpl func(count int) Template
	min  int
	n    int
}

// RandomCount returns a Template that invokes tmpl with a random count read from r
// in [min,max].
//
// If min is equal to max, tmpl is invoked once and the Template returned directly.
//
// An error is returned if either min or max are invalid or outside the suppoted
// range.
func RandomCount(tmpl func(count int) Template, min, max int) (Template, error) {
	if min > max {
		return nil, errors.New("passit: min argument cannot be greater than max argument")
	}

	n := max - min + 1
	if n < 1 || n > maxReadIntN {
		return nil, errors.New("passit: [min,max] range too large")
	}

	if min == max {
		return tmpl(min), nil
	}

	return &randomCount{tmpl, min, n}, nil
}

func (c *randomCount) Password(r io.Reader) (string, error) {
	n, err := readIntN(r, c.n)
	if err != nil {
		return "", err
	}

	return c.tmpl(c.min + n).Password(r)
}

// Space is a Template that always returns a fixed ASCII space.
var Space Template = fixedString(" ")

// Hyphen is a Template that always returns a fixed ASCII hyphen-minus.
var Hyphen Template = fixedString("-")

// FixedString returns a Template that always returns the given string.
func FixedString(s string) Template { return fixedString(s) }

type fixedString string

func (s fixedString) Password(io.Reader) (string, error) { return string(s), nil }
