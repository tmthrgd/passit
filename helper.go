package passit

import (
	"io"
	"slices"
	"strings"
)

type concatGenerator struct {
	gens []Generator
	sep  string
}

// Join returns a Generator that concatenates the outputs of each Generator to
// create a single string. The separator string sep is placed between the outputs in
// the resulting string.
func Join(sep string, gens ...Generator) Generator {
	switch len(gens) {
	case 0:
		return Empty
	case 1:
		return gens[0]
	default:
		return &concatGenerator{slices.Clone(gens), sep}
	}
}

func (cg *concatGenerator) Password(r io.Reader) (string, error) {
	parts := make([]string, len(cg.gens))
	for i, gen := range cg.gens {
		part, err := gen.Password(r)
		if err != nil {
			return "", err
		}

		parts[i] = part
	}

	return strings.Join(parts, cg.sep), nil
}

type repeatGenerator struct {
	gen   Generator
	sep   string
	count int
}

// Repeat returns a Generator that concatenates the output of invoking the Generator
// count times to create a single string. The separator string sep is placed between
// the outputs in the resulting string.
func Repeat(gen Generator, sep string, count int) Generator {
	switch {
	case count < 0:
		panic("passit: count must be positive")
	case count == 0:
		return Empty
	case count == 1:
		return gen
	default:
		return &repeatGenerator{gen, sep, count}
	}
}

func (rg *repeatGenerator) Password(r io.Reader) (string, error) {
	parts := make([]string, rg.count)
	for i := range parts {
		part, err := rg.gen.Password(r)
		if err != nil {
			return "", err
		}

		parts[i] = part
	}

	return strings.Join(parts, rg.sep), nil
}

type randomRepeatGenerator struct {
	gen Generator
	sep string
	min int
	n   int
}

// RandomRepeat returns a Generator that concatenates the output of invoking the
// Generator a random number of times in [min,max] to create a single string. The
// separator string sep is placed between the outputs in the resulting string.
func RandomRepeat(gen Generator, sep string, min, max int) Generator {
	if min < 0 {
		panic("passit: min argument must be positive")
	}
	if min > max {
		panic("passit: min argument cannot be greater than max argument")
	}

	n := max - min + 1
	if n < 1 {
		panic("passit: [min,max] range too large")
	}

	if min == max {
		return Repeat(gen, sep, min)
	}

	return &randomRepeatGenerator{gen, sep, min, n}
}

func (rg *randomRepeatGenerator) Password(r io.Reader) (string, error) {
	n, err := readIntN(r, rg.n)
	if err != nil {
		return "", err
	}

	return Repeat(rg.gen, rg.sep, rg.min+n).Password(r)
}

type alternateGenerator struct {
	gens []Generator
}

// Alternate returns a Generator that randomly selects one of the provided
// Generator's to use to generate the password.
func Alternate(gens ...Generator) Generator {
	switch len(gens) {
	case 0:
		return Empty
	case 1:
		return gens[0]
	default:
		return &alternateGenerator{slices.Clone(gens)}
	}
}

func (ag *alternateGenerator) Password(r io.Reader) (string, error) {
	gen, err := readSliceN(r, ag.gens)
	if err != nil {
		return "", err
	}

	return gen.Password(r)
}

type rejectionGenerator struct {
	gen       Generator
	condition func(string) bool
}

// RejectionSample returns a Generator that continually generates passwords with gen
// until condition reports true for the generated password or an error occurs.
//
// The behaviour is unspecified if condition never reports true.
func RejectionSample(gen Generator, condition func(string) bool) Generator {
	return &rejectionGenerator{gen, condition}
}

func (rg *rejectionGenerator) Password(r io.Reader) (string, error) {
	for {
		pass, err := rg.gen.Password(r)
		if err != nil {
			return "", err
		}
		if rg.condition(pass) {
			return pass, nil
		}
	}
}

type sliceGenerator struct{ list []string }

// FromSlice returns a Generator that returns a random string from list.
func FromSlice(list ...string) Generator {
	switch len(list) {
	case 0:
		return Empty
	case 1:
		return String(list[0])
	default:
		return &sliceGenerator{slices.Clone(list)}
	}
}

func (sg *sliceGenerator) Password(r io.Reader) (string, error) {
	return readSliceN(r, sg.list)
}

type fixedString string

// Empty is a Generator that always returns an empty string.
var Empty Generator = fixedString("")

// Space is a Generator that always returns a fixed ASCII space.
var Space Generator = fixedString(" ")

// Hyphen is a Generator that always returns a fixed ASCII hyphen-minus.
var Hyphen Generator = fixedString("-")

// String returns a Generator that always returns the given string.
func String(s string) Generator {
	return fixedString(s)
}

func (s fixedString) Password(io.Reader) (string, error) {
	return string(s), nil
}
