package passit

import (
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/exp/slices"
)

type words struct{ list []string }

// FromWords returns a Template that returns a random word from list. It returns an
// error if the list of words is invalid.
func FromWords(list ...string) (Template, error) {
	if len(list) < 2 {
		return nil, errors.New("passit: list too short")
	} else if len(list) > maxReadIntN {
		return nil, errors.New("passit: list too long")
	}

	seen := make(map[string]struct{}, len(list))
	for _, word := range list {
		if len(word) < 1 {
			return nil, errors.New("passit: empty word in list")
		} else if !utf8.ValidString(word) {
			return nil, errors.New("passit: word contains invalid unicode rune")
		} else if strings.IndexFunc(word, unicode.IsSpace) >= 0 {
			return nil, errors.New("passit: word contains space")
		}

		if _, dup := seen[word]; dup {
			return nil, errors.New("passit: list contains duplicate word")
		}
		seen[word] = struct{}{}
	}

	return &words{slices.Clone(list)}, nil
}

func (w *words) Password(r io.Reader) (string, error) {
	idx, err := readIntN(r, len(w.list))
	if err != nil {
		return "", err
	}

	return w.list[idx], nil
}
