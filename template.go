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

// Join returns a Template that concatenates the outputs of each Template to create
// a single string. The separator string sep is placed between the outputs in the
// resulting string.
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

type repeated struct {
	tmpl  Template
	sep   string
	count int
}

// Repeat returns a Template that concatenates the output of invoking the Template
// count times to create a single string. The separator string sep is placed between
// the outputs in the resulting string.
func Repeat(t Template, sep string, count int) Template {
	switch {
	case count < 0:
		panic("passit: count must be positive")
	case count == 0:
		return FixedString("")
	case count == 1:
		return t
	default:
		return &repeated{t, sep, count}
	}
}

func (rt *repeated) Password(r io.Reader) (string, error) {
	parts := make([]string, rt.count)
	for i := range parts {
		part, err := rt.tmpl.Password(r)
		if err != nil {
			return "", err
		}

		parts[i] = part
	}

	return strings.Join(parts, rt.sep), nil
}

type randomRepeated struct {
	tmpl Template
	sep  string
	min  int
	n    int
}

// RandomRepeat returns a Template that concatenates the output of invoking the
// Template a random number of times in [min,max] to create a single string. The
// separator string sep is placed between the outputs in the resulting string.
//
// An error is returned if either min or max are invalid or outside the suppoted
// range.
func RandomRepeat(tmpl Template, sep string, min, max int) (Template, error) {
	if min > max {
		return nil, errors.New("passit: min argument cannot be greater than max argument")
	}
	if min < 0 {
		return nil, errors.New("passit: min argument must be positive")
	}

	n := max - min + 1
	if n < 1 || n > maxReadIntN {
		return nil, errors.New("passit: [min,max] range too large")
	}

	if min == max {
		return Repeat(tmpl, sep, min), nil
	}

	return &randomRepeated{tmpl, sep, min, n}, nil
}

func (c *randomRepeated) Password(r io.Reader) (string, error) {
	n, err := readIntN(r, c.n)
	if err != nil {
		return "", err
	}

	return Repeat(c.tmpl, c.sep, c.min+n).Password(r)
}

// Space is a Template that always returns a fixed ASCII space.
var Space Template = fixedString(" ")

// Hyphen is a Template that always returns a fixed ASCII hyphen-minus.
var Hyphen Template = fixedString("-")

// FixedString returns a Template that always returns the given string.
func FixedString(s string) Template { return fixedString(s) }

type fixedString string

func (s fixedString) Password(io.Reader) (string, error) { return string(s), nil }
