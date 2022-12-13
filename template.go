// Package passit provides various password generators.
package passit

import (
	"errors"
	"io"
	"strings"

	"golang.org/x/exp/slices"
)

// Template is an interface for generating passwords.
type Template interface {
	// Password returns a randomly generated password using r as the source of
	// randomness.
	//
	// The returned password may or may not be deterministic with respect to r.
	//
	// r should be a uniformly random stream. The number of bytes read from r
	// may exceed the number of characters in the returned password.
	Password(r io.Reader) (string, error)
}

// The TemplateFunc type is an adapter to allow the use of ordinary functions as
// password generators. If f is a function with the appropriate signature,
// TemplateFunc(f) is a Template that calls f.
type TemplateFunc func(r io.Reader) (string, error)

// Password implements Template, calling f(r).
func (f TemplateFunc) Password(r io.Reader) (string, error) {
	return f(r)
}

type joined struct {
	tmpls []Template
	sep   string
}

// Join returns a Template that concatenates the outputs of each Template to create
// a single string. The separator string sep is placed between the outputs in the
// resulting string.
func Join(sep string, tmpls ...Template) Template {
	switch len(tmpls) {
	case 0:
		return FixedString("")
	case 1:
		return tmpls[0]
	default:
		return &joined{slices.Clone(tmpls), sep}
	}
}

func (j *joined) Password(r io.Reader) (string, error) {
	parts := make([]string, len(j.tmpls))
	for i, tmpl := range j.tmpls {
		part, err := tmpl.Password(r)
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
func Repeat(tmpl Template, sep string, count int) Template {
	switch {
	case count < 0:
		panic("passit: count must be positive")
	case count == 0:
		return FixedString("")
	case count == 1:
		return tmpl
	default:
		return &repeated{tmpl, sep, count}
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

func (rr *randomRepeated) Password(r io.Reader) (string, error) {
	n, err := readIntN(r, rr.n)
	if err != nil {
		return "", err
	}

	return Repeat(rr.tmpl, rr.sep, rr.min+n).Password(r)
}

type alternate struct {
	tmpls []Template
}

// Alternate returns a Template that randomly selects one of the provided Template's
// to use to generate the resultant password.
func Alternate(tmpls ...Template) Template {
	switch len(tmpls) {
	case 0:
		return FixedString("")
	case 1:
		return tmpls[0]
	default:
		return &alternate{slices.Clone(tmpls)}
	}
}

func (at *alternate) Password(r io.Reader) (string, error) {
	n, err := readIntN(r, len(at.tmpls))
	if err != nil {
		return "", err
	}

	return at.tmpls[n].Password(r)
}

type rejection struct {
	tmpl      Template
	condition func(string) bool
}

// RejectionSample returns a Template that continually generates passwords with tmpl
// until condition reports true for the generated password or an error occurs.
//
// The behaviour is unspecified if condition never reports true.
func RejectionSample(tmpl Template, condition func(string) bool) Template {
	return &rejection{tmpl, condition}
}

func (rs *rejection) Password(r io.Reader) (string, error) {
	for {
		pass, err := rs.tmpl.Password(r)
		if err != nil {
			return "", err
		}
		if rs.condition(pass) {
			return pass, nil
		}
	}
}

// Space is a Template that always returns a fixed ASCII space.
var Space Template = fixedString(" ")

// Hyphen is a Template that always returns a fixed ASCII hyphen-minus.
var Hyphen Template = fixedString("-")

// FixedString returns a Template that always returns the given string.
func FixedString(s string) Template {
	return fixedString(s)
}

type fixedString string

func (s fixedString) Password(io.Reader) (string, error) {
	return string(s), nil
}
