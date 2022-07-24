// Package passit provides various password generators.
package passit

import (
	"io"
	"strings"
)

// Template is an interface for generating passwords.
type Template interface {
	// Password returns a randomly generated password using r as the source of
	// randomness.
	Password(r io.Reader) (string, error)
}

type joined struct {
	ts []Template
}

// JoinTemplates returns a Template that returns a password that is the
// concatenation of all the given Templates.
func JoinTemplates(t ...Template) Template {
	if len(t) == 1 {
		return t[0]
	}

	return &joined{append([]Template(nil), t...)}
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

// Space is a Template that always returns a fixed ASCII space.
const Space = space
const space = fixedString(" ")

// FixedString returns a Template that always returns the given string.
func FixedString(s string) Template { return fixedString(s) }

type fixedString string

func (s fixedString) Password(io.Reader) (string, error) { return string(s), nil }
