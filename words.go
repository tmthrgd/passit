package passit

import (
	"errors"
	"io"
	"unicode/utf8"

	"golang.org/x/exp/slices"
)

type words struct{ list []string }

// FromWords returns a Template that returns a random word from list. It returns an
// error if the list of words is invalid.
func FromWords(list ...string) (Template, error) {
	if len(list) > maxReadIntN {
		return nil, errors.New("passit: list too long")
	}

	for _, word := range list {
		if word == "" {
			return nil, errors.New("passit: empty word in list")
		} else if !utf8.ValidString(word) {
			return nil, errors.New("passit: word contains invalid unicode rune")
		}
	}

	switch len(list) {
	case 0:
		return Empty, nil
	case 1:
		return FixedString(list[0]), nil
	default:
		return &words{slices.Clone(list)}, nil
	}
}

func (w *words) Password(r io.Reader) (string, error) {
	idx, err := readIntN(r, len(w.list))
	if err != nil {
		return "", err
	}

	return w.list[idx], nil
}
