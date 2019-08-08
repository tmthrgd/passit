package password

import (
	"io"
	"strings"
)

type Template interface {
	Password(r io.Reader) (string, error)
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
